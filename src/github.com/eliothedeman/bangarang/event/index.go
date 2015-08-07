package event

import (
	"os"
	"sync/atomic"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
)

var (
	EVENT_BUCKET_NAME      = []byte("events")
	INCIDENT_BUCKET_NAME   = []byte("incidents")
	MANAGEMENT_BUCKET_NAME = []byte("management")
	INCIDENT_COUNT_NAME    = []byte("incident_count")
)

const (
	KEEP_ALIVE_SERVICE_NAME = "KeepAlive"
	INDEX_FILE_NAME         = "bnagarang-index.db"
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
	db              *bolt.DB
	incidentCounter *counter
}

func NewIndex() *Index {
	db, err := bolt.Open(INDEX_FILE_NAME, 0600, nil)
	if err != nil {
		logrus.Fatalf("Unable to open index db at %s %s", INDEX_FILE_NAME, err)
	}

	return &Index{
		db:              db,
		incidentCounter: &counter{},
	}
}

// close out the index
func (i *Index) Close() {
	logrus.Info("Closing index")
	err := i.db.Close()
	if err != nil {
		logrus.Errorf("Unable to close index db: %s", err)
	}
}

// delete any psersistants associated with the index
func (i *Index) Delete() {
	logrus.Info("Deleting index")
	i.Close()
	err := os.Remove(INDEX_FILE_NAME)
	if err != nil {
		logrus.Errorf("Unable to delete the index db: %s", err)
	}
}

func (i *Index) GetIncidentCounter() int64 {
	return i.incidentCounter.get()
}

func (i *Index) UpdateIncidentCounter(count int64) {
	i.incidentCounter.set(count)
}

// write the incident to the db
func (i *Index) PutIncident(in *Incident) {
	i.db.Update(func(tx *bolt.Tx) {
		b := tx.Bucket(INCIDENT_BUCKET_NAME)
		b.Put(in.IndexName(), value)
	})
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
