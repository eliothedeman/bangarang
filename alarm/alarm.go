package alarm

import (
	"encoding/json"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/event"
	"github.com/eliothedeman/workhorse/config"
)

var (
	alarms = &alarmCollection{
		// maps an alarm type name to an alarm
		factories:   make(map[string]AlarmFactory),
		collections: AlarmCollection{},
	}
)

// AlarmCollection maps the name of an escalation policy to the actions to be taken by them
type AlarmCollection map[string][]Alarm

// UnmarshalJSON a custom unmarshal func for an escalation policy
func (a AlarmCollection) UnmarshalJSON(buff []byte) error {
	b := map[string][]json.RawMessage{}
	name := &struct {
		Type *string `json:"type"`
		Name *string `json:"name"`
	}{}

	err := json.Unmarshal(buff, &b)
	if err != nil {
		return err
	}

	for k, v := range b {
		a[k] = make([]Alarm, 0)
		for _, raw := range v {
			name.Name = nil
			name.Type = nil
			json.Unmarshal(raw, name)

			fact := GetFactory(*name.Type)
			new_alarm := fact()
			conf := new_alarm.ConfigStruct()
			json.Unmarshal(raw, conf)
			err := new_alarm.Init(conf)
			if err != nil {
				return err
			}
			a[k] = append(a[k], new_alarm)

		}
		LoadCollection(k, a[k])
	}

	return nil
}

func LoadCollection(name string, coll []Alarm) {
	alarms.Lock()
	alarms.collections[name] = coll
	alarms.Unlock()
}

func GetCollection(name string) []Alarm {
	alarms.Lock()
	a, ok := alarms.collections[name]
	alarms.Unlock()
	if !ok {
		return nil
	}
	return a
}
func GetFactory(name string) AlarmFactory {
	alarms.Lock()
	a := alarms.factories[name]
	alarms.Unlock()
	return a
}

func LoadFactory(name string, f AlarmFactory) {
	logrus.Infof("Loading alarm factory %s", name)
	alarms.Lock()
	alarms.factories[name] = f
	alarms.Unlock()
}

type Alarm interface {
	Send(i *event.Incident) error
	config.Configer
}

type AlarmFactory func() Alarm

type alarmCollection struct {
	collections AlarmCollection
	factories   map[string]AlarmFactory
	sync.Mutex
}
