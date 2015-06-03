package event

import (
	"encoding/binary"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/pquerna/ffjson/ffjson"
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

type Index struct {
	pool       *EncodingPool
	db         *bolt.DB
	dbFileName string
	keepAlives map[string]time.Time
	ka_lock    sync.RWMutex
}

func NewIndex(dbName string) *Index {
	db_wait := make(chan struct{})
	var db *bolt.DB
	var err error
	go func() {
		db, err = bolt.Open(dbName, 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		db_wait <- struct{}{}
	}()

	// if the db takes more than 100 miliseconds to open, fail out
	select {
	case <-time.After(100 * time.Millisecond):
		log.Fatalf("Unable to open db %s", dbName)
	case <-db_wait:
		log.Printf("Opened db %s", dbName)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(EVENT_BUCKET_NAME)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(MANAGEMENT_BUCKET_NAME)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(INCIDENT_BUCKET_NAME)
		return err
	})

	if err != nil {
		log.Fatal(err)
	}

	return &Index{
		pool:       NewEncodingPool(NewMsgPackEncoder, NewMsgPackDecoder, runtime.NumCPU()),
		db:         db,
		dbFileName: dbName,
		keepAlives: make(map[string]time.Time),
	}
}

// close out the index
func (i *Index) Close() error {
	i.ka_lock.Lock()
	i.keepAlives = make(map[string]time.Time)
	i.ka_lock.Unlock()
	return i.db.Close()
}

// delete any psersistants associated with the index
func (i *Index) Delete() error {
	err := i.Close()
	if err != nil {
		log.Println(err)
	}

	return os.Remove(i.dbFileName)
}

// return all keep alive's as events
func (i *Index) GetKeepAlives() []*Event {
	n := time.Now()
	i.ka_lock.Lock()
	events := make([]*Event, len(i.keepAlives))
	x := 0
	for host, t := range i.keepAlives {
		events[x] = &Event{
			Host:    host,
			Metric:  n.Sub(t).Seconds(),
			Service: KEEP_ALIVE_SERVICE_NAME,
		}
		x += 1
	}
	i.ka_lock.Unlock()
	return events
}

func (i *Index) GetIncidentCounter() int64 {
	var buff []byte

	i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MANAGEMENT_BUCKET_NAME)
		buff = b.Get(INCIDENT_COUNT_NAME)
		return nil
	})

	// if the counter wasn't found, set it to 0 and return the value
	if len(buff) == 0 {
		i.UpdateIncidentCounter(0)
		return 0
	}

	if len(buff) != 8 {
		log.Println("incorrect size of counter buffer")
		i.UpdateIncidentCounter(0)
		return 0
	}

	count, _ := binary.Varint(buff)
	return count
}

func (i *Index) UpdateIncidentCounter(count int64) {
	buff := make([]byte, 8)
	binary.PutVarint(buff, count)

	i.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(MANAGEMENT_BUCKET_NAME)
		b.Put(INCIDENT_COUNT_NAME, buff)
		return nil
	})
}

// write the incident to the db
func (i *Index) PutIncident(in *Incident) {
	buff, err := ffjson.MarshalFast(in)
	if err != nil {
		log.Println(err)
		return
	}

	err = i.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(INCIDENT_BUCKET_NAME)
		return b.Put(in.IndexName(), buff)
	})

	if err != nil {
		log.Println(err)
	}

	return
}

// list all the known events
func (i *Index) ListIncidents() []*Incident {
	var ins []*Incident

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(INCIDENT_BUCKET_NAME)

		// create a slice large enough to hole every value of the incidents bucket
		ins = make([]*Incident, 0, b.Stats().KeyN)

		// for every incident, parse it and add it to the incidents collection
		err := b.ForEach(func(k, v []byte) error {
			in := &Incident{}
			err := ffjson.UnmarshalFast(v, in)
			if err != nil {
				return err
			}

			ins = append(ins, in)
			return nil
		})

		return err
	})

	if err != nil {
		log.Println(err)
	}

	return ins
}

// get an event from the index
func (i *Index) GetIncident(id []byte) *Incident {
	var buff []byte

	// attempt to find the incident in the index
	i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(INCIDENT_BUCKET_NAME)
		buff = b.Get(id)
		return nil
	})

	// if we couldn't find the incident
	if len(buff) == 0 {
		return nil
	}

	in := &Incident{}
	// if we have the event, attempt to decode it
	err := ffjson.Unmarshal(buff, in)

	if err != nil {
		log.Println(err)
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
		log.Println(err)
	}
}

// updates the event, will not apply any of the dedupe logic
func (i *Index) UpdateEvent(e *Event) {
	var buff []byte
	var err error
	i.pool.Encode(func(enc Encoder) {
		buff, err = enc.Encode(e)
	})
	if err != nil {
		log.Println(err)
		return
	}

	err = i.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(EVENT_BUCKET_NAME)
		return b.Put([]byte(e.IndexName()), buff)
	})

	if err != nil {
		log.Println(err)
		return
	}

}

// insert the event into the index
func (i *Index) PutEvent(e *Event) {

	// update the host's keepalive value
	i.ka_lock.Lock()
	i.keepAlives[e.Host] = time.Now()
	i.ka_lock.Unlock()
}
