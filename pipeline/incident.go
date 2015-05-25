package pipeline

import (
	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/event"
)

//go:generate ffjson $GOFILE

// An incident is created whenever an event changes to a state that is not event.OK
type Incident struct {
	Event      *event.Event      `json:"event"`
	Time       int64             `json:"time"`
	Id         int               `json:"id"`
	Active     bool              `json:"active"`
	Escalation *alarm.Escalation `json:"escalation"`
}
