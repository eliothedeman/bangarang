package test

import (
	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/event"
)

type TestAlert struct {
	Events map[*event.Event]int
}

func init() {
	alarm.LoadFactory("test", NewTest)
}

type Console struct {
}

func (t *TestAlert) Send(e *event.Event) error {
	t.Events[e] = e.Status
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
