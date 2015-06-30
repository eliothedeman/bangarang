package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
)

// DBConf provides methods for reading and writing configs from a database
type DBConf struct {
	fileName string
	db       *bolt.DB
}

const (
	// CurrentBucketName is the name of the bucket that holds the current
	// verison of the config
	CurrentBucketName = "current"

	// CurrentVersionHash is the "hash" name of the current config
	CurrentVersionHash  = "current"
	appConfigBucketName = "app"
)

var (
	// BucketNames holds the names of the buckets used to store versioned configs
	BucketNames = []string{
		"app",
	}
)

func (d *DBConf) initBuckets() {
	err := d.db.Update(func(tx *bolt.Tx) error {
		for _, name := range BucketNames {
			_, err := tx.CreateBucketIfNotExists([]byte(name))
			if err != nil {
				return err
			}

		}
		return nil
	})

	if err != nil {
		logrus.Errorf("Unable to init config buckets %s", err)
	}
}

func createIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f, err := os.Create(path)
		f.Close()
		return err
	}

	return nil
}

func (d *DBConf) initDB() error {
	createIfNotExists(d.fileName)
	db, err := bolt.Open(d.fileName, 0600, nil)
	d.db = db
	d.initBuckets()

	return err
}

func (d *DBConf) encode(i interface{}) ([]byte, error) {
	return json.Marshal(i)
}

func (d *DBConf) decode(buff []byte, i interface{}) error {
	return json.Unmarshal(buff, i)
}

// Snapshot represents a bangarang config at a given point in time
type Snapshot struct {
	Hash      string
	TimeStamp time.Time
	App       *AppConfig
}

func newSnapshot(ac *AppConfig) *Snapshot {
	return &Snapshot{
		TimeStamp: time.Now(),
		App:       ac,
		Hash:      fmt.Sprintf("%x", HashConfig(ac)),
	}
}

func (d *DBConf) getVersion(version string) (*AppConfig, error) {
	logrus.Infof("Loading config version %s", version)

	var buff []byte
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(appConfigBucketName))
		if b == nil {
			var err error
			b, err = tx.CreateBucket([]byte(appConfigBucketName))
			if err != nil {
				return err
			}
		}
		buff = b.Get([]byte(version))
		return nil
	})

	if err != nil {
		return nil, err
	}

	// decode the snapshot
	s := newSnapshot(NewDefaultConfig())
	s.App.provider = d

	// if the buffer is of zero size, then the config was not found
	if len(buff) == 0 {
		logrus.Infof("No config found for version %s in db, using defaults", version)
		return s.App, nil
	}

	err = d.decode(buff, s)

	if err != nil {
		return nil, err
	}

	// compile all of the policies
	for _, p := range s.App.Policies {
		p.Compile()
	}
	s.App.provider = d

	return s.App, nil
}

// GetCurrent loads the current version of the config
func (d *DBConf) GetCurrent() (*AppConfig, error) {
	return d.getVersion(CurrentVersionHash)
}

// GetConfig get the config file which has the hash of given version
func (d *DBConf) GetConfig(versionHash string) (*AppConfig, error) {
	return d.getVersion(versionHash)
}

// PutConfig writes the given config to the database and returns
// the new hash and an error
func (d *DBConf) PutConfig(a *AppConfig) (string, error) {
	err := d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(appConfigBucketName))
		oldBuff := b.Get([]byte(CurrentVersionHash))
		if len(oldBuff) > 0 {
			old := newSnapshot(NewDefaultConfig())
			err := d.decode(oldBuff, old)
			if err != nil {
				log.Println(err)
				return err
			}

			// write the old snapshot at it's
			err = b.Put([]byte(old.Hash), oldBuff)
			if err != nil {
				log.Println(err)
				return err
			}
		}

		// write the new snapshot to disk
		newBuff, err := d.encode(newSnapshot(a))
		if err != nil {
			log.Println(err)
			return err
		}

		return b.Put([]byte(CurrentVersionHash), newBuff)
	})

	if err != nil {
		log.Println(err)
		return "", err
	}

	return string(a.Hash), nil
}
