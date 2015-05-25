package event

import (
	"encoding/binary"
	"time"
)

//go:generate ffjson $GOFILE

// An incident is created whenever an event changes to a state that is not event.OK
type Incident struct {
	EventName  string `json:"event"`
	Time       int64  `json:"time"`
	Id         int64  `json:"id"`
	Active     bool   `json:"active"`
	Status     int    `json:"status"`
	Escalation string `json:"escalation"`
}

func (i *Incident) IndexName() []byte {
	buff := make([]byte, 8)
	binary.PutVarint(buff, i.Id)
	return buff
}
func NewIncident(escalation string, i *Index, e *Event) *Incident {
	i.UpdateIncidentCounter(i.GetIncidentCounter() + 1)
	in := &Incident{
		EventName:  e.IndexName(),
		Time:       time.Now().Unix(),
		Active:     true,
		Status:     e.Status,
		Escalation: escalation,
		Id:         i.GetIncidentCounter(),
	}

	// add the incident to the event, so we can look it up later
	e.Incident = in

	i.UpdateEvent(e)
	i.PutIncident(in)

	return in
}
