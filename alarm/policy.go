package alarm

import (
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/eliothedeman/bangarang/event"
)

func init() {
	log.SetFlags(log.Llongfile)
}

type Policy struct {
	Match       map[string]string `json:"match"`
	NotMatch    map[string]string `json:"not_match"`
	Crit        *Condition        `json:"crit"`
	Warn        *Condition        `json:"warn"`
	Name        string            `json:"name"`
	r_match     map[string]*regexp.Regexp
	r_not_match map[string]*regexp.Regexp
}

// check to see if an event satisfies the policy
func (p *Policy) Matches(e *event.Event) bool {
	return p.CheckMatch(e) && p.CheckNotMatch(e)
}

// compile the regex patterns for this policy
func (p *Policy) Compile() {
	if p.r_match == nil {
		p.r_match = make(map[string]*regexp.Regexp)
	}

	if p.r_not_match == nil {
		p.r_not_match = make(map[string]*regexp.Regexp)
	}

	for k, v := range p.Match {
		p.r_match[k] = regexp.MustCompile(v)
	}

	for k, v := range p.NotMatch {
		p.r_not_match[k] = regexp.MustCompile(v)
	}

	if p.Crit != nil {
		p.Crit.init()
	}

	if p.Warn != nil {
		p.Warn.init()
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
func (p *Policy) Action(e *event.Event) string {
	if p.Crit != nil {
		if p.Crit.TrackEvent(e) {
			e.Status = event.CRITICAL
			return p.Crit.Escalation
		}
	}
	if p.Warn != nil {
		if p.Warn.TrackEvent(e) {
			e.Status = event.WARNING
			return p.Warn.Escalation
		}
	}

	e.Status = event.OK
	return ""
}

func (p *Policy) CheckNotMatch(e *event.Event) bool {
	v := reflect.ValueOf(e).Elem()
	for k, m := range p.r_not_match {
		elem := v.FieldByName(formatFileName(k))
		if m.MatchString(elem.String()) {
			return false

			// check againt the element's tags
			if e.Tags != nil {
				if against, inMap := e.Tags[k]; inMap {
					if m.MatchString(against) {
						return false
					}
				}
			}
		}
	}

	return true
}

// check if any of p's matches are satisfied by the event
func (p *Policy) CheckMatch(e *event.Event) bool {
	v := reflect.ValueOf(e).Elem()
	for k, m := range p.r_match {
		elem := v.FieldByName(formatFileName(k))

		// if the element does not match the regex pattern, the event does not fully match
		if !m.MatchString(elem.String()) {

			// check againt the element's tags
			if e.Tags == nil {
				return false
			}
			if against, inMap := e.Tags[k]; inMap {
				if !m.MatchString(against) {
					return false
				}
			} else {
				return false
			}
		}
	}

	return true
}
