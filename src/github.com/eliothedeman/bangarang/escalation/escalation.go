package escalation

import (
	"encoding/json"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/event"
)

var (
	Escalations = &EscalationCollection{
		// maps an Escalation type name to an Escalation
		factories: make(map[string]Factory),
	}
)

// EscalationPolicy is the collection of escalations that are subscribed to by the policy
type EscalationPolicy struct {
	Match    *event.TagSet     `json:"match"`
	NotMatch *event.TagSet     `json:"not_match"`
	Warn     bool              `json:"warn"`
	Crit     bool              `json:"crit"`
	Ok       bool              `json:"ok"`
	Name     string            `json:"name"`
	Comment  string            `json:"comment"`
	Configs  []json.RawMessage `json:"configs"`

	// compiled regex matches
	rMatch    Matcher
	rNotMatch Matcher

	// escalations to forward incidents to
	escalations []Escalation
}

// Compile sets up all the regex matches for the subscriptions and starts all of the escalations held by the policy
func (e *EscalationPolicy) Compile() (err error) {
	// create a matcher from each tagset
	e.rMatch, err = MatcherFromTagSet(e.Match)
	if err != nil {
		return
	}
	e.rNotMatch, err = MatcherFromTagSet(e.NotMatch)
	if err != nil {
		return
	}

	// create enough space for all of the new escalations
	e.escalations = make([]Escalation, 0, len(e.Configs))

	// go through each config and creat an escalation out of it
	for _, raw := range e.Configs {

		// run parsing logic on the config
		newEscalation, perr := parseEscalation(raw)
		if perr != nil {
			return perr
		}

		// if all is well, append the new escalation
		e.escalations = append(e.escalations, newEscalation)

	}

	return nil
}

// isSubscribed will return true if this policy is subscribed to incidents like the one given
func (e *EscalationPolicy) isSubscribed(i *event.Incident) bool {

	// check that the policy is subscribed to the incident's status
	switch i.Status {
	case event.OK:
		if !e.Ok {
			return false
		}
	case event.WARNING:
		if !e.Warn {
			return false
		}
	case event.CRITICAL:
		if !e.Crit {
			return false
		}
	default:
		// don't forward any bad status
		return false
	}

	return e.rMatch.MatchesAll(i.Tags) && !e.rNotMatch.MatchesOne(i.Tags)
}

// Pass an incident into the escalation for processing
func (e *EscalationPolicy) Pass(i *event.Incident) {

	// only process incidents that this policy subscribes to
	if e.isSubscribed(i) {

		// send if off to every escalation known about
		for _, ep := range e.escalations {
			err := ep.Send(i)
			if err != nil {
				logrus.Errorf("Unable to forward incident %s to escalation %+v", i.FormatDescription(), ep)
			}
		}
	}
}

// parseEscalation given a raw config, create a new escalation
func parseEscalation(buff json.RawMessage) (Escalation, error) {
	name := &struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}{}

	var err error

	// parse out the name/escalation type
	err = json.Unmarshal(buff, name)
	if err != nil {
		return nil, err
	}

	// create a new escalation of the correct type
	newEscalation := GetFactory(name.Type)()

	// get the config struct to unmarshal into
	conf := newEscalation.ConfigStruct()

	// unmarshal into config struct
	err = json.Unmarshal(buff, conf)
	if err != nil {
		return nil, err
	}

	// init the new escalation with the config
	err = newEscalation.Init(conf)
	return newEscalation, err
}

// GetFactory returns the Factory associated with the given name
func GetFactory(name string) Factory {
	Escalations.Lock()
	a := Escalations.factories[name]
	Escalations.Unlock()
	return a
}

// LoadFactory loads an Factory into the globaly available map of EscalationFactories
func LoadFactory(name string, f Factory) {
	logrus.Debugf("Loading Escalation factory %s", name)
	Escalations.Lock()
	Escalations.factories[name] = f
	Escalations.Unlock()
}

// Escalation is the basic interface which provides a way to communicate incidents
// to the outside world
type Escalation interface {
	Send(i *event.Incident) error
	ConfigStruct() interface{}
	Init(interface{}) error
}

// Factory returns a new Escalation
type Factory func() Escalation

type EscalationCollection struct {
	factories map[string]Factory
	sync.Mutex
}
