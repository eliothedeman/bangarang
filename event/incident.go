package event

import "time"

//go:generate ffjson $GOFILE
//go:generate msgp $GOFILE

// An incident is created whenever an event changes to a state that is not event.OK
type Incident struct {
	EventName   []byte `json:"event" msg:"event_name"`
	Time        int64  `json:"time" msg:"time"`
	Id          int64  `json:"id" msg:"id"`
	Active      bool   `json:"active" msg:"active"`
	Escalation  string `json:"escalation" msg:"escalation"`
	Description string `json:"description" msg:"description"`
	Policy      string `json:"policy" msg:"policy"`
	Event
}

func (i *Incident) IndexName() []byte {
	return []byte(i.Policy + i.Event.IndexName())
}

func (i *Incident) GetEvent() *Event {
	return &i.Event
}

func (i *Incident) FormatDescription() string {
	return i.Description
}

func NewIncident(policy string, e *Event) *Incident {
	in := &Incident{
		EventName:   []byte(e.IndexName()),
		Time:        time.Now().Unix(),
		Active:      true,
		Policy:      policy,
		Description: e.FormatDescription(),
		Event:       *e,
	}

	return in
}
