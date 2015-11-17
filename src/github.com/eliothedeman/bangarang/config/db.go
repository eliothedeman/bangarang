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
	CurrentVersionHash   = "current"
	appConfigBucketName  = "app"
	userConfigBucketName = "user"
)

var (
	// BucketNames holds the names of the buckets used to store versioned configs
	BucketNames = []string{
		"app",
		"user",
	}
)

func (d *DBConf) initAdminUser() error {
	adminExists := false
	err := d.db.Update(func(t *bolt.Tx) error {
		b, err := t.CreateBucketIfNotExists([]byte(userConfigBucketName))
		if err != nil {
			return err
		}

		// the admin user is always user_id 0
		buff := b.Get([]byte("admin"))

		if len(buff) != 0 {
			adminExists = true
		}

		return nil
	})

	// if the admin already exists, no need to do anything
	if adminExists {
		return nil
	}

	if err != nil {
		return err
	}

	// if no user is found, we must created one
	u := NewUser("admin", "admin", "admin", ADMIN)
	logrus.Info("Adding default admin user. user: admin passs: admin")
	// add the user to the database
	return d.PutUser(u)
}

func (d *DBConf) init() {

	// get the schema, and apply it before moving on
	s := GetSchemaFromDb(d.db)

	// bootstrap
	if s.Version == First {
		logrus.Info("Bootstrapping config db")
		err := s.Apply(d.db)
		if err != nil {
			logrus.Errorf("Unable to bootstrap config %s", err.Error())
		}

	}

	for LatestSchema().Greater(GetSchemaFromDb(d.db)) {
		logrus.Infof("Upgrading config db version from %s to %s", s.Version, LatestSchema().Version)
		err := LatestSchema().Apply(d.db)
		if err != nil {
			logrus.Errorf("Unable to apply schema version %s to config db %s", LatestSchema().Version, err.Error())
		} else {
			s = LatestSchema()
		}
	}

	logrus.Infof("Using db config version %s", s.Version)

	err := d.initAdminUser()
	if err != nil {
		logrus.Errorf("Unable init admin user: %s", err)
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
	d.init()

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
	Hash            string     `json:"hash"`
	Timestamp       time.Time  `json:"time_stamp"`
	App             *AppConfig `json:"app"`
	CreatorId       uint16     `json:"creator_id"` // the User.Id of who created this snapshot
	CreatorName     string     `json:"creator_name"`
	CreatorUserName string     `json:"creator_user_name"`
}

func newSnapshot(ac *AppConfig, creator *User) *Snapshot {
	return &Snapshot{
		Timestamp:       time.Now(),
		App:             ac,
		CreatorName:     creator.Name,
		CreatorUserName: creator.UserName,
		Hash:            fmt.Sprintf("%x", HashConfig(ac)),
	}
}

//  GetUserByUserName
func (d *DBConf) GetUserByUserName(name string) (*User, error) {

	// get all the users
	users, err := d.ListUsers()
	if err != nil {
		return nil, err
	}

	// for every user, check to see if it has the user name we are looking for
	for _, u := range users {
		if u.UserName == name {
			return u, nil
		}
	}

	return nil, fmt.Errorf("Unable to find users with name %s", name)
}

//  GetUser by their User.Id
func (d *DBConf) GetUser(name string) (*User, error) {
	var buff []byte

	err := d.db.View(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(userConfigBucketName))
		buff = b.Get([]byte(name))
		return nil
	})

	if err != nil {
		return nil, err
	}

	// if the found buffer is of len 0, then the user's record was not found
	if len(buff) == 0 {
		return nil, fmt.Errorf("User: %s not found", name)
	}

	// unmarshal the user
	u := &User{}
	err = d.decode(buff, u)

	return u, err
}

// PutUser inserts the user into the db
func (d *DBConf) PutUser(u *User) error {

	// encode the user
	buff, err := d.encode(u)
	if err != nil {
		return err
	}

	// write the user to the db
	err = d.db.Update(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(userConfigBucketName))
		return b.Put([]byte(u.UserName), buff)
	})

	return err
}

// DeleteUser by the User.Id
func (d *DBConf) DeleteUser(name string) error {
	return d.db.Update(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(userConfigBucketName))
		return b.Delete([]byte(name))
	})
}

// ListUsers fetches all known users
func (d *DBConf) ListUsers() ([]*User, error) {
	var u []*User

	err := d.db.View(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(userConfigBucketName))
		u = make([]*User, 0, b.Stats().KeyN)

		// for every key/value decode the user and append it to the user list
		return b.ForEach(func(key, val []byte) error {
			x := &User{}
			err := d.decode(val, x)
			if err != nil {
				return err
			}

			// add the user to the list
			u = append(u, x)
			return nil
		})

	})
	return u, err
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
	s := &Snapshot{
		App: NewDefaultConfig(),
	}
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

func (d *DBConf) ListRawSnapshots() []json.RawMessage {
	raw := []json.RawMessage{}
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(appConfigBucketName))
		return b.ForEach(func(k, v []byte) error {
			raw = append(raw, json.RawMessage(v))
			return nil
		})
	})

	if err != nil {
		logrus.Error(err)
	}

	return raw

}

func (d *DBConf) ListSnapshots() []*Snapshot {
	snaps := []*Snapshot{}
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(appConfigBucketName))
		return b.ForEach(func(k, v []byte) error {
			s := &Snapshot{}
			err := d.decode(v, s)
			if err != nil {
				return err
			}
			s.Hash = string(k)
			snaps = append(snaps, s)
			return nil
		})
	})
	if err != nil {
		logrus.Error(err)
	}
	return snaps
}

// PutConfig writes the given config to the database and returns
// the new hash and an error
func (d *DBConf) PutConfig(a *AppConfig, u *User) (string, error) {

	// check to see if the given user has write permissions
	if u.Permissions < WRITE {
		return "", InsufficientPermissions(WRITE, u.Permissions)
	}
	err := d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(appConfigBucketName))
		oldBuff := b.Get([]byte(CurrentVersionHash))
		if len(oldBuff) > 0 {
			old := newSnapshot(NewDefaultConfig(), u)
			err := d.decode(oldBuff, old)
			if err != nil {
				log.Println(err)
				return err
			}

			// write the old snapshot at it's hash
			err = b.Put([]byte(old.Hash), oldBuff)
			if err != nil {
				log.Println(err)
				return err
			}
		}

		// write the new snapshot to disk
		newBuff, err := d.encode(newSnapshot(a, u))
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
