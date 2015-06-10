package event

import (
	"sync"
	"sync/atomic"

	"github.com/Sirupsen/logrus"
)

var (
	EVENT_BUCKET_NAME      = []byte("events")
	INCIDENT_BUCKET_NAME   = []byte("incidents")
	MANAGEMENT_BUCKET_NAME = []byte("management")
	INCIDENT_COUNT_NAME    = []byte("incident_count")
)

const (
	KEEP_ALIVE_SERVICE_NAME = "KeepAlive"
)

type counter struct {
	c int64
}

func (c *counter) inc() {
	atomic.AddInt64(&c.c, 1)
}
func (c *counter) set(val int64) {
	atomic.StoreInt64(&c.c, val)
}

func (c *counter) get() int64 {
	return atomic.LoadInt64(&c.c)
}

type Index struct {
	i_lock          sync.RWMutex
	incidents       map[string]Incident
	incidentCounter *counter
}

func NewIndex() *Index {
	return &Index{
		incidents:       make(map[string]Incident),
		incidentCounter: &counter{},
	}
}

// close out the index
func (i *Index) Close() {
	logrus.Info("Closing index")
	i.incidentCounter.set(0)
	i.i_lock.Lock()
	i.incidents = make(map[string]Incident)
	i.i_lock.Unlock()

}

// delete any psersistants associated with the index
func (i *Index) Delete() {
	logrus.Info("Deleting index")
}

func (i *Index) GetIncidentCounter() int64 {
	return i.incidentCounter.get()
}

func (i *Index) UpdateIncidentCounter(count int64) {
	i.incidentCounter.set(count)
}

// write the incident to the db
func (i *Index) PutIncident(in *Incident) {
	i.i_lock.Lock()
	i.incidents[string(in.IndexName())] = *in
	i.i_lock.Unlock()
}

// list all the known events
func (i *Index) ListIncidents() []*Incident {
	i.i_lock.RLock()
	ins := make([]*Incident, len(i.incidents))
	x := 0
	for _, v := range i.incidents {
		ins[x] = &v
		x += 1
	}
	i.i_lock.RUnlock()
	return ins
}

// get an event from the index
func (i *Index) GetIncident(id []byte) *Incident {
	i.i_lock.RLock()
	defer i.i_lock.RUnlock()

	in, ok := i.incidents[string(id)]
	if !ok {
		return nil
	}
	return &in
}

func (i *Index) DeleteIncidentById(id []byte) {
	i.i_lock.Lock()
	delete(i.incidents, string(id))
	i.i_lock.Unlock()
}
