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

// Collection maps the name of an escalation policy to the actions to be taken by them
type Collection struct {
	Coll map[string][]Escalation
	raw  map[string][]json.RawMessage
}

// Collection holds a map of strings to Escalations
func (c *Collection) Collection() map[string][]Escalation {
	if c.Coll == nil {
		c.Coll = map[string][]Escalation{}
	}
	return c.Coll
}

// AddRaw map's a raw escalation to it's name
func (c *Collection) AddRaw(name string, raw []json.RawMessage) {
	if c.raw == nil {
		c.raw = make(map[string][]json.RawMessage)
	}
	c.raw[name] = raw
}

// RemoveRaw removes a raw value from teh collection if it exists
func (c *Collection) RemoveRaw(name string) {
	delete(c.raw, name)
}

// MarshalJSON encode the collection as json
func (c *Collection) MarshalJSON() ([]byte, error) {
	return json.Marshal(&c.raw)
}

// UnmarshalRaw Runs unmarshaling logic over the raw values stored in the collection
func (c *Collection) UnmarshalRaw() error {
	name := &struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}{}
	c.Coll = make(map[string][]Escalation)
	var err error
	for k, v := range c.raw {
		c.Coll[k] = make([]Escalation, 0)
		for _, raw := range v {
			name.Name = ""
			name.Type = ""
			err = json.Unmarshal(raw, name)
			if err != nil {
				return err
			}

			fact := GetFactory(name.Type)
			newEscalation := fact()
			conf := newEscalation.ConfigStruct()
			err = json.Unmarshal(raw, conf)
			if err != nil {
				return err
			}
			err = newEscalation.Init(conf)
			if err != nil {
				return err
			}
			c.Coll[k] = append(c.Coll[k], newEscalation)
		}
	}

	return nil
}

// UnmarshalJSON a custom unmarshal func for an escalation policy
func (c *Collection) UnmarshalJSON(buff []byte) error {
	c.raw = make(map[string][]json.RawMessage)
	err := json.Unmarshal(buff, &c.raw)
	if err != nil {
		return err
	}
	return c.UnmarshalRaw()
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
	collections Collection
	factories   map[string]Factory
	sync.Mutex
}
