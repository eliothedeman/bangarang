package event

import (
	"crypto/md5"
	"fmt"
	"time"
)

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
	Status      int    `json:"status" "msg:"status"`
	indexName   []byte
	Event
}

func (i *Incident) IndexName() []byte {
	if len(i.indexName) == 0 {
		n := md5.New()
		n.Write([]byte(i.Policy + i.Escalation + i.Event.IndexName()))
		i.indexName = []byte(fmt.Sprintf("%x", n.Sum(nil)))
	}

	return i.indexName
}

func (i *Incident) GetEvent() *Event {
	return &i.Event
}

func (i *Incident) FormatDescription() string {
	return fmt.Sprintf("%s on %s is %s. Triggerd by %s", i.Service, i.Host, status(i.Status), i.Policy)
}

func NewIncident(policy string, escalation string, status int, e Event) *Incident {
	in := &Incident{
		EventName:  []byte(e.IndexName()),
		Time:       time.Now().Unix(),
		Active:     true,
		Status:     status,
		Policy:     policy,
		Escalation: escalation,
		Event:      e,
	}
	in.Description = in.FormatDescription()

	return in
}
