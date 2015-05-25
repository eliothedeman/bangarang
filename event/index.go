package event

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/pquerna/ffjson/ffjson"
)

var (
	EVENT_BUCKET_NAME = []byte("events")
)

type Index struct {
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

	db.NoSync = true

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(EVENT_BUCKET_NAME)
		return err
	})

	if err != nil {
		log.Fatal(err)
	}

	return &Index{
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

// get all of the hosts that have missed their keepalives
func (i *Index) GetExpired(age time.Duration) []string {
	hosts := make([]string, 0, 10)
	n := time.Now()
	i.ka_lock.Lock()
	for host, t := range i.keepAlives {
		if n.Sub(t) > age {
			hosts = append(hosts, host)
		}
	}
	i.ka_lock.Unlock()
	return hosts
}

// updates the event, will not apply any of the dedupe logic
func (i *Index) UpdateEvent(e *Event) {
	buff, err := ffjson.MarshalFast(e)
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
	name := []byte(e.IndexName())
	e.LastEvent = i.GetEvent(name)
	if e.LastEvent != nil {
		e.LastEvent.LastEvent = nil
	}

	// encode the event
	buff, err := ffjson.MarshalFast(e)
	if err != nil {
		log.Println(err)
		return
	}

	// write the event to the db
	err = i.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(EVENT_BUCKET_NAME)
		return b.Put(name, buff)
	})

	ffjson.Pool(buff)

	// update the host's keepalive value
	i.ka_lock.Lock()
	i.keepAlives[e.Host] = time.Now()
	i.ka_lock.Unlock()
}

// fetch the event with the given index name
func (i *Index) GetEvent(name []byte) *Event {
	e := &Event{}
	found := false
	err := i.db.View(func(t *bolt.Tx) error {
		b := t.Bucket(EVENT_BUCKET_NAME)
		raw := b.Get(name)

		// if we don't have anything at that key, exit early
		if len(raw) == 0 {
			return nil
		}
		found = true
		err := ffjson.UnmarshalFast(raw, e)

		return err
	})
	if err != nil {
		log.Println(err)
		return nil
	}

	if !found {
		return nil
	}

	return e
}
