package alarm

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/event"
)

func init() {
	log.SetFlags(log.Llongfile)
}

var (
	DEFAULT_GROUP_BY = map[string]string{
		"host":        "^(.*)$",
		"service":     "^(.*)$",
		"sub_service": "^(.*)$",
	}
)

type Policy struct {
	Match       map[string]string `json:"match"`
	NotMatch    map[string]string `json:"not_match"`
	GroupBy     map[string]string `json:"group_by"`
	Crit        *Condition        `json:"crit"`
	Warn        *Condition        `json:"warn"`
	Name        string            `json:"name"`
	Comment     string            `json:"comment"`
	r_match     map[string]*regexp.Regexp
	r_not_match map[string]*regexp.Regexp
	stop        chan struct{}
	in          chan *pack
	resolve     chan *event.Incident
}

// start the policy listening for events
func (p *Policy) start() {
	go func() {
		var in *pack
		for {
			select {
			case toResolve := <-p.resolve:
				var c *Condition

				// fetch the condition that created this incident
				if toResolve.Status == event.CRITICAL {
					c = p.Crit
				} else if toResolve.Status == event.WARNING {
					c = p.Warn
				}

				if c != nil {
					e := &event.Event{}
					e.Host = toResolve.Host
					e.Service = toResolve.Service
					e.SubService = toResolve.SubService
					t := c.getTracker(e)
					t.refresh()
				}
			case <-p.stop:
				logrus.Info("Stopping policy", p.Name)

				// cleanup the policy. Sometimes they hangaround.
				p.clean()

				// stop this policy from being stopped again
				p.stop = nil

				// catch resolve's that could be sent to this policy
				res := p.resolve
				p.resolve = nil

				go func() {
					for {
						select {
						case <-res:
							logrus.Info("Attempted to resolve an incident on a policy that no longer exists")

						case <-time.After(1 * time.Minute):
							return
						}
					}

				}()
				return
			case in = <-p.in:

				// process the event if it matches the policy
				if p.Matches(in.e) {
					// process the request

					// check critical
					if shouldAlert, status := p.ActionCrit(in.e); shouldAlert {
						incident := event.NewIncident(p.Name, p.Crit.Escalation, status, in.e)
						incident.SetResolve(p.resolve)

						in.n(incident)
						logrus.Info(incident.FormatDescription())

						// check warning
					} else if shouldAlert, status := p.ActionWarn(in.e); shouldAlert {
						incident := event.NewIncident(p.Name, p.Warn.Escalation, status, in.e)
						incident.SetResolve(p.resolve)
						in.n(incident)
						logrus.Info(incident.FormatDescription())
					}
				}
				in.e.WaitDec()
			}
		}
	}()
}

// remove all used memeory by this policy
func (p *Policy) clean() {
	p.Crit = nil
	p.Warn = nil
	p.r_match = nil
	p.r_not_match = nil
}

type pack struct {
	e *event.Event
	n func(*event.Incident)
}

// Process will send exicute the next function if the event satisfies the policy
func (p *Policy) Process(e *event.Event, next func(*event.Incident)) {
	p.in <- &pack{
		e: e,
		n: next,
	}
}

func (p *Policy) Stop() {
	p.stop <- struct{}{}
}

// check to see if an event satisfies the policy
func (p *Policy) Matches(e *event.Event) bool {
	return p.CheckMatch(e) && p.CheckNotMatch(e)
}

// compile the regex patterns for this policy
func (p *Policy) Compile() {
	logrus.Infof("Compiling regex maches for %s", p.Name)
	p.in = make(chan *pack, 10)
	p.stop = make(chan struct{})
	p.resolve = make(chan *event.Incident)
	p.start()

	if p.r_match == nil {
		p.r_match = make(map[string]*regexp.Regexp)
	}

	if p.r_not_match == nil {
		p.r_not_match = make(map[string]*regexp.Regexp)
	}

	// if we don't have at least three componants of the group by, establish them from the defaults
	if len(p.GroupBy) < 3 {

		if len(p.GroupBy) == 0 {
			p.GroupBy = DEFAULT_GROUP_BY

		} else {

			tmp := map[string]string{}
			for k, v := range DEFAULT_GROUP_BY {
				tmp[k] = v
			}

			for k, v := range p.GroupBy {
				tmp[k] = v
			}

			p.GroupBy = tmp
		}
	}

	for k, v := range p.Match {

		m, err := regexp.Compile(v)
		if err != nil {
			logrus.Errorf("Unable to compile match for %s: %s", k, err.Error())
		} else {
			p.r_match[k] = m
		}
	}

	for k, v := range p.NotMatch {
		m, err := regexp.Compile(v)
		if err != nil {
			logrus.Errorf("Unable to compile not_match for %s: %s", k, err.Error())
		} else {
			p.r_not_match[k] = m
		}
	}

	if p.Crit != nil {
		logrus.Infof("Initializing crit for %s", p.Name)
		p.Crit.init(p.GroupBy)
	}

	if p.Warn != nil {
		logrus.Infof("Initializing warn for %s", p.Name)
		p.Warn.init(p.GroupBy)
	}
}

func formatFileName(n string) string {
	s := strings.Split(n, "_")
	a := ""
	for _, k := range s {
		a = a + strings.Title(k)
	}
	return a
}

// return the action to take for a given event
func (p *Policy) ActionCrit(e *event.Event) (bool, int) {
	status := event.OK
	if p.Crit != nil {
		if p.Crit.TrackEvent(e) {
			status = event.CRITICAL
		} else {
			status = event.OK
		}

		return p.Crit.StateChanged(e), status
	}

	return false, status
}

func (p *Policy) ActionWarn(e *event.Event) (bool, int) {
	status := event.OK
	if p.Warn != nil {
		if p.Warn.TrackEvent(e) {
			status = event.WARNING
		} else {
			status = event.OK
		}
		return p.Warn.StateChanged(e), status
	}

	return false, status
}

func (p *Policy) CheckNotMatch(e *event.Event) bool {
	for k, m := range p.r_not_match {
		if m.MatchString(e.Get(k)) {
			return false
		}
	}
	return true
}

// check if any of p's matches are satisfied by the event
func (p *Policy) CheckMatch(e *event.Event) bool {
	for k, m := range p.r_match {
		// if the element does not match the regex pattern, the event does not fully match
		if !m.MatchString(e.Get(k)) {
			return false
		}
	}
	return true
}
