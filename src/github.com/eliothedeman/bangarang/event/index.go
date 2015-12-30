package event

import (
	"encoding/json"
	"fmt"
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
	INDEX_FILE_NAME        = "bangarang-index.db"
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
	db              *bolt.DB
	incidentCounter *counter
}

func NewIndex() *Index {
	db, err := bolt.Open(INDEX_FILE_NAME, 0600, nil)
	if err != nil {
		logrus.Fatalf("Unable to open index db at %s %s", INDEX_FILE_NAME, err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(INCIDENT_BUCKET_NAME)
		return err
	})
	if err != nil {
		logrus.Fatal("Unable to create incident bucket in the index db")
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
	err := i.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(INCIDENT_BUCKET_NAME)
		buff, err := json.Marshal(in)
		if err != nil {
			return err
		}
		return b.Put(in.IndexName(), buff)
	})
	if err != nil {
		logrus.Errorf("Unable to insert incident into index %s", err)
	}
}

// list all the known events
func (i *Index) ListIncidents() []*Incident {
	var ins []*Incident
	err := i.db.View(func(t *bolt.Tx) error {
		b := t.Bucket(INCIDENT_BUCKET_NAME)
		ins = make([]*Incident, b.Stats().KeyN)
		x := 0
		return b.ForEach(func(k, v []byte) error {
			y := &Incident{}
			err := json.Unmarshal(v, y)
			if err != nil {
				logrus.Warnf("Invalid json: key=%s val=%s", string(k), string(v))
				ins[x] = y
				return err
			}

			ins[x] = y
			x++
			return nil
		})
	})

	if err != nil {
		logrus.Errorf("Unable to list incidents %s", err)
	}
	return ins
}

// get an event from the index
func (i *Index) GetIncident(id []byte) *Incident {
	in := &Incident{}
	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(INCIDENT_BUCKET_NAME)
		buff := b.Get(id)
		if len(buff) == 0 {
			return fmt.Errorf("Unable to find incident with id %s", string(id))
		}

		err := json.Unmarshal(buff, in)
		return err
	})

	if err != nil {
		return nil
	}

	return in
}

func (i *Index) DeleteIncidentById(id []byte) {
	err := i.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(INCIDENT_BUCKET_NAME)
		return b.Delete(id)
	})

	if err != nil {
		logrus.Errorf("Unable to delete incident %s from index", string(id))
	}
}
