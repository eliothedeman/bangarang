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

type Escalation struct {
	Policy Policy
	Alarms []Alarm
}

func (e *Escalation) Match(ev *event.Event) bool {
	return e.Policy.match(ev) && e.Policy.matchNots(ev)
}

func (e *Escalation) StatusOf(ev *event.Event) int {
	return e.Policy.StatusOf(ev)
}

type Policy struct {
	Match       map[string]string `match`
	NotMatch    map[string]string `not_match`
	Crit        *Condition        `crit`
	Warn        *Condition        `warn`
	r_match     map[string]*regexp.Regexp
	r_not_match map[string]*regexp.Regexp
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
}

func formatFiledName(n string) string {
	s := strings.Split(n, "_")
	a := ""
	for _, k := range s {
		a = a + strings.Title(k)
	}
	return a
}

func (p *Policy) StatusOf(e *event.Event) int {
	if p.Crit != nil {
		if p.Crit.Satisfies(e) {
			if p.Crit.TrackEvent(e) {
				e.Status = event.CRITICAL
				return event.CRITICAL
			}
		}
		p.Crit.CleanEvent(e)
	}
	if p.Warn != nil {
		if p.Warn.Satisfies(e) {
			if p.Warn.TrackEvent(e) {
				e.Status = event.WARNING
				return event.WARNING
			}
		}
		p.Warn.CleanEvent(e)
	}

	e.Status = event.OK
	return event.OK
}

func (p *Policy) matchNots(e *event.Event) bool {
	v := reflect.ValueOf(e).Elem()
	for k, m := range p.r_not_match {
		elem := v.FieldByName(formatFiledName(k))
		if m.MatchString(elem.String()) {
			return false
		}
	}

	return true
}

// check if any of p's matches are satisfied by the event
func (p *Policy) match(e *event.Event) bool {
	v := reflect.ValueOf(e).Elem()
	for k, m := range p.r_match {
		elem := v.FieldByName(formatFiledName(k))
		if m.MatchString(elem.String()) {
			return true
		}
	}

	return false
}
