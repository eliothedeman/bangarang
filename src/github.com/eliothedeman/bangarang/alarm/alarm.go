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
	c   map[string][]Alarm
	raw map[string][]json.RawMessage
}

func (c *Collection) Collection() map[string][]Alarm {
	return c.c
}

func (c *Collection) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.raw)
}

// UnmarshalJSON a custom unmarshal func for an escalation policy
func (c *Collection) UnmarshalJSON(buff []byte) error {
	b := map[string][]json.RawMessage{}
	name := &struct {
		Type *string `json:"type"`
		Name *string `json:"name"`
	}{}

	err := json.Unmarshal(buff, &b)
	if err != nil {
		return err
	}

	c.raw = make(map[string][]json.RawMessage)
	c.c = make(map[string][]Alarm)
	for k, v := range b {
		c.c[k] = make([]Alarm, 0)
		c.raw[k] = v
		for _, raw := range v {
			name.Name = nil
			name.Type = nil
			json.Unmarshal(raw, name)

			fact := GetFactory(*name.Type)
			newAlarm := fact()
			conf := newAlarm.ConfigStruct()
			json.Unmarshal(raw, conf)
			err := newAlarm.Init(conf)
			if err != nil {
				return err
			}
			c.c[k] = append(c.c[k], newAlarm)

		}
	}

	return nil
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
