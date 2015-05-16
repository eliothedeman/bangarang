package alarm

import (
	"sync"

	"github.com/eliothedeman/bangarang/event"
	"github.com/eliothedeman/workhorse/config"
)

var (
	alarms = &alarmCollection{
		// maps an instance name to an alarm
		alarms: make(map[string]Alarm),
		// maps an alarm type name to an alarm
		factories: make(map[string]AlarmFactory),
	}
)

func LoadFactory(name string, f AlarmFactory) {
	alarms.Lock()
	alarms.factories[name] = f
	alarms.Unlock()
}

type Alarm interface {
	Send(e *event.Event) error
	config.Configer
}

type AlarmFactory func() Alarm

type alarmCollection struct {
	alarms    map[string]Alarm
	factories map[string]AlarmFactory
	sync.Mutex
}
