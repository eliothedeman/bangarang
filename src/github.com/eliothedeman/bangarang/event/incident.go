package event

import (
	"crypto/md5"
	"fmt"
	"time"
)

// IncidentPasser passes an incdent to the next step in the pipeline
type IncidentPasser interface {
	PassIncident(i *Incident)
}

// IncidentFormatter will createa a string representation of an incident
type IncidentFormatter func(*Incident) string

// DefaultIncidentFormatter is the legacy formatter for incidents
func DefaultIncidentFormatter(i *Incident) string {
	return fmt.Sprintf("%s on %s is %s. Triggered by %s", i.Tags.Get("service"), i.Tags.Get("host"), Status(i.Status), i.Policy)
}

// An incident is created whenever an event changes to a state
type Incident struct {
	EventName   []byte `json:"event" msg:"event_name"`
	Time        int64  `json:"time" msg:"time"`
	Id          int64  `json:"id" msg:"id"`
	Active      bool   `json:"active" msg:"active"`
	Description string `json:"description" msg:"description"`
	Policy      string `json:"policy" msg:"policy"`
	Status      int    `json:"status" "msg:"status"`
	indexName   []byte
	resChan     chan *Incident // this is used to call back to the policy that created this event
	Event
}

// SetResolve sets the incident resolver channel for the given incident
func (i *Incident) SetResolve(r chan *Incident) {
	i.resChan = r
}

// GetResolve return the incidents resover channel
func (i *Incident) GetResolve() chan *Incident {
	return i.resChan
}

// IndexName returns the unique name for an incident of this description
func (i *Incident) IndexName() []byte {
	if len(i.indexName) == 0 {
		n := md5.New()
		n.Write([]byte(i.Policy + i.Event.Tags.String()))
		i.indexName = []byte(fmt.Sprintf("%x", n.Sum(nil)))
	}

	return i.indexName
}

// GetEvent returns a pointer to the event housed within the incident
func (i *Incident) GetEvent() *Event {
	return &i.Event
}

// FormatDescription calls the formatter for this incident
func (i *Incident) FormatDescription() string {
	return DefaultIncidentFormatter(i)
}

// NewIncident creats and returns a new incident for the given event
func NewIncident(policy string, status int, e *Event) *Incident {
	in := &Incident{
		EventName: []byte(e.IndexName()),
		Time:      time.Now().Unix(),
		Active:    true,
		Status:    status,
		Policy:    policy,
		Event:     *e,
	}

	in.Tags = e.Tags
	in.Metric = e.Metric
	in.Description = in.FormatDescription()

	return in
}
