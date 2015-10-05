package version

import "github.com/boltdb/bolt"

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

func (s Schema) Newer(x Schema) bool {
	return s.Version.Newer(x.Version)
}

func GetSchemaFromDb(b *bolt.DB) Schema {

	// first version
	v := First
	err := b.View(func(t *bolt.Tx) error {
		b := t.Bucket(AboutBucket)
		if b == nil {
			return nil
		}

		v = VersionFromString(string(b.Get([]byte("about"))))
		return nil
	})
	if err != nil {
		return Schemas[0]
	}

	// find the oldest valid schema
	for _, s := range Schemas {

		// best case, exact match
		if s.Version == v {
			return s
		}

		// as soon as we find an older version, this is the one
		if !s.Version.Newer(v) {
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
