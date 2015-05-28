package event

import (
	"encoding/binary"
	"time"
)

//go:generate ffjson $GOFILE
//go:generate msgp $GOFILE

// An incident is created whenever an event changes to a state that is not event.OK
type Incident struct {
	EventName  []byte `json:"event" msg:"event_name"`
	Time       int64  `json:"time" msg:"time"`
	Id         int64  `json:"id" msg:"id"`
	Active     bool   `json:"active" msg:"active"`
	Status     int    `json:"status" msg:"status"`
	Escalation string `json:"escalation" msg:"escalation"`
}

func (i *Incident) IndexName() []byte {
	buff := make([]byte, 8)
	binary.PutVarint(buff, i.Id)
	return buff
}
func NewIncident(escalation string, i *Index, e *Event) *Incident {
	i.UpdateIncidentCounter(i.GetIncidentCounter() + 1)
	in := &Incident{
		EventName:  []byte(e.IndexName()),
		Time:       time.Now().Unix(),
		Active:     true,
		Status:     e.Status,
		Escalation: escalation,
		Id:         i.GetIncidentCounter(),
	}

	// add the incident to the event, so we can look it up later
	id := in.Id
	e.IncidentId = &id

	i.UpdateEvent(e)
	i.PutIncident(in)

	return in
}
