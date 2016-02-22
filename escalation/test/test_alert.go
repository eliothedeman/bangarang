package test

import (
	"sync"

	"github.com/eliothedeman/bangarang/escalation"
	"github.com/eliothedeman/bangarang/event"
)

type TestAlert struct {
	Incidents []*event.Incident
	sync.Mutex
}

func init() {
	escalation.LoadFactory("test", NewTest)
}

func (t *TestAlert) Do(f func(*TestAlert)) {
	t.Lock()
	f(t)
	t.Unlock()
}

func (t *TestAlert) Send(i *event.Incident) error {
	t.Do(func(t *TestAlert) {
		t.Incidents = append(t.Incidents, i)
	})
	return nil
}

func (t *TestAlert) ConfigStruct() interface{} {
	return &struct{}{}
}

func (t *TestAlert) Init(i interface{}) error {
	return nil
}

func NewTest() escalation.Escalation {
	return &TestAlert{
		Incidents: make([]*event.Incident, 0),
	}
}

func NewTestAlert() *TestAlert {
	return &TestAlert{
		Incidents: make([]*event.Incident, 0),
	}
}
