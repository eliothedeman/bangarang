package alarm

import (
	"encoding/json"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/event"
)

var (
	alarms = &alarmCollection{
		// maps an alarm type name to an alarm
		factories: make(map[string]Factory),
	}
)

// Collection maps the name of an escalation policy to the actions to be taken by them
type Collection struct {
	Coll map[string][]Alarm
	raw  map[string][]json.RawMessage
}

// Collection holds a map of strings to alarms
func (c *Collection) Collection() map[string][]Alarm {
	if c.Coll == nil {
		c.Coll = map[string][]Alarm{}
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
	c.Coll = make(map[string][]Alarm)
	var err error
	for k, v := range c.raw {
		c.Coll[k] = make([]Alarm, 0)
		for _, raw := range v {
			name.Name = ""
			name.Type = ""
			err = json.Unmarshal(raw, name)
			if err != nil {
				return err
			}

			fact := GetFactory(name.Type)
			newAlarm := fact()
			conf := newAlarm.ConfigStruct()
			err = json.Unmarshal(raw, conf)
			if err != nil {
				return err
			}
			err = newAlarm.Init(conf)
			if err != nil {
				return err
			}
			c.Coll[k] = append(c.Coll[k], newAlarm)
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
	alarms.Lock()
	a := alarms.factories[name]
	alarms.Unlock()
	return a
}

// LoadFactory loads an Factory into the globaly available map of AlarmFactories
func LoadFactory(name string, f Factory) {
	logrus.Debugf("Loading alarm factory %s", name)
	alarms.Lock()
	alarms.factories[name] = f
	alarms.Unlock()
}

// Alarm is the basic interface which provides a way to communicate incidents
// to the outside world
type Alarm interface {
	Send(i *event.Incident) error
	ConfigStruct() interface{}
	Init(interface{}) error
}

// Factory returns a new alarm
type Factory func() Alarm

type alarmCollection struct {
	collections Collection
	factories   map[string]Factory
	sync.Mutex
}
