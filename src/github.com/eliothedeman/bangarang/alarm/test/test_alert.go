package test

import (
	"sync"

	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/event"
)

type TestAlert struct {
	Events map[*event.Event]int
	sync.Mutex
}

func init() {
	alarm.LoadFactory("test", NewTest)
}

func (t *TestAlert) Do(f func(*TestAlert)) {
	t.Lock()
	f(t)
	t.Unlock()
}

func (t *TestAlert) Send(i *event.Incident) error {
	t.Do(func(t *TestAlert) {
		t.Events[i.GetEvent()] = i.Status
	})
	return nil
}

func (t *TestAlert) ConfigStruct() interface{} {
	return &struct{}{}
}

func (t *TestAlert) Init(i interface{}) error {
	return nil
}

func NewTest() alarm.Alarm {
	return &TestAlert{
		Events: make(map[*event.Event]int),
	}
}
