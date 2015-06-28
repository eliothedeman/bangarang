package config

import (
	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
)

type DBConf struct {
	fileName string
	db       *bolt.DB
}

var (
	BUCKET_NAMES = []string{
		"master",
		"providers",
		"alerts",
		"ui",
	}
	CURRENT_BUCKET_NAME = []byte("current")
)

func (d *DBConf) initBuckets() {
	err := d.db.Update(func(tx *bolt.Tx) error {
		for _, name := range BUCKET_NAMES {
			b, err := tx.CreateBucketIfNotExists([]byte(name))
			if err != nil {
				return err
			}

			// create an inner bucket for "current" versions of all of these
			_, err = b.CreateBucketIfNotExists(CURRENT_BUCKET_NAME)
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

func (d *DBConf) initDB() error {
	db, err := bolt.Open(d.fileName, 0600, nil)
	d.db = db
	d.initBuckets()

	return err
}

func (d *DBConf) encode(i interface{}) ([]byte, error) {
	return json.Marshal(i)
}

func (d *DBConf) decode(buff []byte, i interface{}) error {

}

func (d *DBConf) GetCurrent() (*AppConfig, error) {
	ac := NewDefaultConfig()

	// load up the master config

}

func (d *DBConf) GetConfig() (*AppConfig, error) {
	ac := NewDefaultConfig()
	return ac, nil
}
