package config

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

const (
	sqlite = "sqlite3"
)

// SQLConf is a config provider which uses a sqlite3 database as it's persistence
type SQLConf struct {
	db gorm.DB
}

// Attempts to connect to the sqlite3 database
func (s *SQLConf) connect(path string) error {
	db, err := gorm.Open(sqlite3DBType, path)
	s.db = db
	return err
}

// NewSQLConf inits a new config provider backed by a sql database
func NewSQLConf(path string) (Provider, error) {
	s := &SQLConf{
		path: path,
	}

	return s, s.connect()
}
