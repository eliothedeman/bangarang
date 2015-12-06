package escalation

import (
	"log"
	"regexp"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/event"
)

func init() {
	log.SetFlags(log.Llongfile)
}

var (
	DEFAULT_GROUP_BY = &event.TagSet{
		{"host", "^(.*)$"},
		{"service", "^(.*)$"},
		{"sub_service", "^(.*)$"},
	}
)

type Policy struct {
	Match       *event.TagSet `json:"match"`
	NotMatch    *event.TagSet `json:"not_match"`
	GroupBy     *event.TagSet `json:"group_by"`
	Crit        *Condition    `json:"crit"`
	Warn        *Condition    `json:"warn"`
	Name        string        `json:"name"`
	Comment     string        `json:"comment"`
	next        event.IncidentPasser
	r_match     Matcher
	r_not_match Matcher
	stop        chan struct{}
	in          chan *event.Event
	resolve     chan *event.Incident
}

// start the policy listening for events
func (p *Policy) start() {
	go func() {
		var e *event.Event
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
					t := c.getTracker(&toResolve.Event)
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
			case e = <-p.in:

				// process the event if it matches the policy

				if p.Matches(e) {

					// check critical
					if shouldAlert, status := p.ActionCrit(e); shouldAlert {
						incident := event.NewIncident(p.Name, status, e)
						incident.SetResolve(p.resolve)

						// send send it off to the next hop
						incident.GetEvent().WaitInc()
						p.next.PassIncident(incident)

						// check warning
					} else if shouldAlert, status := p.ActionWarn(e); shouldAlert {
						incident := event.NewIncident(p.Name, status, e)
						incident.SetResolve(p.resolve)

						// send it off to the next hop
						incident.GetEvent().WaitInc()
						p.next.PassIncident(incident)
					}
				}
				e.WaitDec()
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

// Process will send exicute the next function if the event satisfies the policy
func (p *Policy) PassEvent(e *event.Event) {
	p.in <- e
}

func (p *Policy) Stop() {
	p.stop <- struct{}{}
}

// check to see if an event satisfies the policy
func (p *Policy) Matches(e *event.Event) bool {

	// if there is nothing to positivly match on, there is no match
	if p.Match.Len() == 0 {
		return false
	}

	return p.CheckMatch(e) && !p.CheckNotMatch(e)
}

type Matcher []MatchSet

type MatchSet struct {
	Key   string
	Value *regexp.Regexp
}

func MatcherFromTagSet(t *event.TagSet) (Matcher, error) {
	m := make(Matcher, len(*t))
	i := 0
	var verr error
	t.ForEach(func(k, v string) {
		match, err := regexp.Compile(v)
		if err != nil {
			verr = err
		} else {
			m[i] = MatchSet{
				Key:   k,
				Value: match,
			}
		}
	})
	return m, verr
}

// MatchesOne returns true if the matcher matches at least one item in the TagSet
func (m Matcher) MatchesOne(t *event.TagSet) (matches bool) {
	m.ForEach(func(k string, v *regexp.Regexp) {
		if v.MatchString(t.Get(k)) {
			matches = true
			return
		}
	})

	return
}

// MatchesAll returns true if the TagSet satisfies the entire matcher
func (m *Matcher) MatchesAll(t *event.TagSet) (matches bool) {
	matches = true
	m.ForEach(func(k string, v *regexp.Regexp) {
		if !v.MatchString(t.Get(k)) {
			matches = false
		}
	})
	return

}

func (m Matcher) ForEach(f func(k string, v *regexp.Regexp)) {
	for _, t := range m {
		f(t.Key, t.Value)
	}
}

// compile the regex patterns for this policy
func (p *Policy) Compile(next event.IncidentPasser) {
	logrus.Infof("Compiling regex maches for %s", p.Name)
	p.in = make(chan *event.Event, 10)
	p.stop = make(chan struct{})
	p.resolve = make(chan *event.Incident)
	p.next = next
	p.start()

	if p.r_match == nil {

		// handle the nil case
		if p.Match != nil {
			p.r_match = make(Matcher, len(*p.Match))

		} else {
			p.r_match = make(Matcher, 0)
		}
	}

	if p.r_not_match == nil {
		if p.NotMatch != nil {
			p.r_not_match = make(Matcher, len(*p.NotMatch))

		} else {
			p.r_not_match = make(Matcher, 0)
		}
	}

	i := 0
	p.Match.ForEach(func(k, v string) {
		m, err := regexp.Compile(v)
		if err != nil {
			logrus.Errorf("Unable to compile match for %s: %s", k, err.Error())
		} else {
			p.r_match[i] = MatchSet{
				Key:   k,
				Value: m,
			}
		}
		i += 1
	})

	p.NotMatch.ForEach(func(k, v string) {
		m, err := regexp.Compile(v)
		if err != nil {
			logrus.Errorf("Unable to compile not_match for %s: %s", k, err.Error())
		} else {
			p.r_not_match[i] = MatchSet{
				Key:   k,
				Value: m,
			}
		}
		i += 1
	})

	if p.Crit != nil {
		logrus.Infof("Initializing crit for %s", p.Name)
		p.Crit.init(p.GroupBy)
	}

	if p.Warn != nil {
		logrus.Infof("Initializing warn for %s", p.Name)
		p.Warn.init(p.GroupBy)
	}
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

// CheckNotMatch returns true if any of the not_match's are satisfied by the TagSet
func (p *Policy) CheckNotMatch(e *event.Event) bool {
	return p.r_not_match.MatchesOne(e.Tags)
}

// CheckMatch returns true if all of match's are satisfied by the event's TagSet
func (p *Policy) CheckMatch(e *event.Event) bool {
	return p.r_match.MatchesAll(e.Tags)
}
