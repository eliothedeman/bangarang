package config

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
)

var (
	AboutBucket = []byte("about")
	Schemas     = []Schema{
		{
			Version: Version{
				0, 10, 4,
			},
			Buckets: []string{
				"app",
				"user",
			},

			// A no-op. This is the first schemad version
			Upgrader: func(old *bolt.DB) error {
				return nil
			},
		},
		{
			Version: Version{
				0, 12, 0,
			},
			Buckets: []string{
				"app",
				"user",
			},
			Upgrader: func(old *bolt.DB) error {

				upgradeRawSnapshot := func(buff []byte) ([]byte, error) {
					// old
					oldSnap := make(map[string]interface{})
					err := json.Unmarshal(buff, &oldSnap)
					if err != nil {
						return nil, err
					}

					// remove the incompatible parts and reencode
					// TODO (eliothedeman) actually attempt to update semi-compatible things

					app, ok := oldSnap["app"].(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("Unable to upgrade snapshot. No 'app' section found.")
					}

					// walk through the policies, and remove the match/not_match fields
					raw_pols, ok := app["policies"]
					if ok {
						pols, ok := raw_pols.(map[string]interface{})
						if ok {
							for k, v := range pols {

								logrus.Debugf("Removing old match/not_match from %s", k)
								v.(map[string]interface{})["match"] = []struct{}{}
								v.(map[string]interface{})["not_match"] = []struct{}{}
								v.(map[string]interface{})["group_by"] = []struct{}{}
							}
						} else {
							log.Println(reflect.TypeOf(raw_pols).String())
						}

					}

					// delete the global policy
					app["global_policy"] = map[string]interface{}{}

					// add it back to the old snapshot
					oldSnap["app"] = app

					// reencode
					return json.Marshal(&oldSnap)

				}

				// open for an update
				err := old.Update(func(tx *bolt.Tx) error {

					// get the bucket for app snapshots
					b := tx.Bucket([]byte("app"))
					if b == nil {
						logrus.Warning("No config snapshots found")
						return nil
					}

					// make something that can hold all of the new snapshots to be written after the upgrade
					newSnapshots := make([]struct {
						key []byte
						val []byte
					}, 0, b.Stats().KeyN)

					b.ForEach(func(k, v []byte) error {
						logrus.Debugf("Starting snapshot upgrade on: %s", string(k))
						upgraded, err := upgradeRawSnapshot(v)
						if err != nil {
							logrus.Errorf("Unable to upgrade snapshot: %s %s", string(k), err.Error())
						}

						newSnapshots = append(newSnapshots, struct {
							key []byte
							val []byte
						}{
							key: k,
							val: upgraded,
						})

						return nil
					})

					// updated every snapshot with the upgraded config
					for _, snap := range newSnapshots {
						err := b.Put(snap.key, snap.val)
						if err != nil {
							logrus.Errorf("Unable to write upgraded snapshot to database %s %s", string(snap.key), err.Error())
						}

						logrus.Debugf("Finished snapshot upgrade on: %s", string(snap.key))
					}

					return nil

				})

				return err

			},
		},
		{
			Version: Version{
				0, 13, 0,
			},
			Buckets: []string{
				"app",
				"user",
			},
			Upgrader: func(old *bolt.DB) error {
				upgradeRawSnapshot := func(buff []byte) ([]byte, error) {
					oldSnap := make(map[string]interface{})
					err := json.Unmarshal(buff, &oldSnap)
					if err != nil {
						return nil, err
					}

					app, ok := oldSnap["app"].(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("Unable to upgrade snapshot. No 'app' section found.")
					}

					// remove the escalations
					app["escalations"] = map[string]interface{}{}

					// reencode the message
					return json.Marshal(&app)
				}
				// open for an update
				err := old.Update(func(tx *bolt.Tx) error {

					// get the bucket for app snapshots
					b := tx.Bucket([]byte("app"))
					if b == nil {
						logrus.Warning("No config snapshots found")
						return nil
					}

					// make something that can hold all of the new snapshots to be written after the upgrade
					newSnapshots := make([]struct {
						key []byte
						val []byte
					}, 0, b.Stats().KeyN)

					b.ForEach(func(k, v []byte) error {
						logrus.Debugf("Starting snapshot upgrade on: %s", string(k))
						upgraded, err := upgradeRawSnapshot(v)
						if err != nil {
							logrus.Errorf("Unable to upgrade snapshot: %s %s", string(k), err.Error())
						}

						newSnapshots = append(newSnapshots, struct {
							key []byte
							val []byte
						}{
							key: k,
							val: upgraded,
						})

						return nil
					})

					// updated every snapshot with the upgraded config
					for _, snap := range newSnapshots {
						err := b.Put(snap.key, snap.val)
						if err != nil {
							logrus.Errorf("Unable to write upgraded snapshot to database %s %s", string(snap.key), err.Error())
						}

						logrus.Debugf("Finished snapshot upgrade on: %s", string(snap.key))
					}

					return nil

				})

				return err

			},
		},
	}
)

// LatestSchema returns the newest schema
func LatestSchema() Schema {
	return Schemas[len(Schemas)-1]
}

// givena n old and new schema update the current db
type Upgrader func(old *bolt.DB) error

// Schema represents a bucket schema for a bolt.Db
type Schema struct {
	Version  Version
	Buckets  []string
	Upgrader Upgrader
}

// Greater
func (s Schema) Greater(x Schema) bool {
	return s.Version.Greater(x.Version)
}

func GetSchemaFromDb(b *bolt.DB) Schema {

	// first version
	v := First
	err := b.View(func(t *bolt.Tx) error {
		b := t.Bucket(AboutBucket)
		if b == nil {
			return nil
		}

		v = VersionFromString(string(b.Get([]byte("version"))))
		log.Println(v)
		return nil
	})
	if err != nil {
		return Schemas[0]
	}

	// find the oldest valid schema
	for i := len(Schemas) - 1; i >= 0; i-- {
		s := Schemas[i]

		// best case, exact match
		if s.Version.String() == v.String() {
			return s
		}

		// as soon as we find an older version, this is the one
		if !s.Version.Greater(v) {
			return s
		}
	}

	return Schemas[0]
}

// Apply creates all needed buckets and sets the version of the db
func (s Schema) Apply(b *bolt.DB) error {

	// make sure all buckets that are needed for the upgrade are in place
	err := s.createBuckets(b)
	if err != nil {
		return err
	}

	// run the upgrader
	err = s.Upgrader(b)
	if err != nil {
		return err
	}

	// update the "about" section of the database
	err = s.putAbout(b)
	if err != nil {
		return err
	}

	// clean up
	return s.cleanBuckets(b)

}

// cleanup uneeded buckets
func (s Schema) cleanBuckets(b *bolt.DB) error {

	isNeeded := func(name string) bool {

		if name == string(AboutBucket) {
			return true
		}

		for _, k := range s.Buckets {
			if k == name {
				return true
			}
		}

		return false

	}
	return b.Update(func(t *bolt.Tx) error {

		buckets := [][]byte{}

		// find all the nill buckets
		err := t.ForEach(func(k []byte, _ *bolt.Bucket) error {
			buckets = append(buckets, k)
			return nil
		})

		if err != nil {
			return err
		}

		// for every buckets, check to see if it is still needed
		for _, name := range buckets {

			if !isNeeded(string(name)) {
				err = t.DeleteBucket([]byte(name))
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (s Schema) createBuckets(b *bolt.DB) error {
	return b.Update(func(t *bolt.Tx) error {
		for _, bucket := range s.Buckets {
			_, err := t.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (s Schema) putAbout(b *bolt.DB) error {
	return b.Update(func(t *bolt.Tx) error {
		b, err := t.CreateBucketIfNotExists(AboutBucket)
		if err != nil {
			return err
		}

		return b.Put([]byte("version"), []byte(s.Version.String()))
	})
}
