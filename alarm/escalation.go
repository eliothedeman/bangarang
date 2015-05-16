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

type Policy struct {
	Match       map[string]string `match`
	NotMatch    map[string]string `not_match`
	Crit        *Condition        `crit`
	Warn        *Condition        `warn`
	Ok          *Condition        `ok`
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

// check if any of p's matches are satisfied by the event
func (p *Policy) MatchAny(e *event.Event) bool {
	v := reflect.ValueOf(e).Elem()

	// check the not matches first
	for k, m := range p.r_not_match {
		elem := v.FieldByName(formatFiledName(k))
		if m.MatchString(elem.String()) {
			return false
		}
	}

	for k, m := range p.r_match {
		elem := v.FieldByName(formatFiledName(k))
		if m.MatchString(elem.String()) {
			return true
		}
	}

	return false
}

type Condition struct {
	Greater    *float64 `greater`
	Less       *float64 `less`
	Exactly    *float64 `exactly`
	Occurences *int     `occurences`
}
