package event

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/eliothedeman/smoothie"
)

//go:generate ffjson $GOFILE
//go:generate msgp $GOFILE

// An incident is created whenever an event changes to a state that is not event.OK
type Incident struct {
	EventName   []byte    `json:"event" msg:"event_name"`
	Time        int64     `json:"time" msg:"time"`
	Id          int64     `json:"id" msg:"id"`
	Active      bool      `json:"active" msg:"active"`
	Escalation  string    `json:"escalation" msg:"escalation"`
	Description string    `json:"description" msg:"description"`
	Policy      string    `json:"policy" msg:"policy"`
	GraphURL    string    `json:"graph_url" msg:"graph_url"`
	GraphData   []float64 `json:"graph_data" msg:"graph_data"`
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
	return i.Event.FormatDescription()
}

func NewIncident(policy, escalation string, e *Event, context *smoothie.DataFrame) *Incident {
	in := &Incident{
		EventName:   []byte(e.IndexName()),
		Time:        time.Now().Unix(),
		Active:      true,
		Policy:      policy,
		Escalation:  escalation,
		Description: e.FormatDescription(),
		Event:       *e,
		GraphData:   context.Data(),
	}

	in.GraphURL = fmt.Sprintf("api/incident/graph/%s", string(in.IndexName()))

	return in
}
