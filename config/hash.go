package config

import (
	"crypto/md5"
	"encoding/json"

	"github.com/Sirupsen/logrus"
)

func HashConfig(c *AppConfig) []byte {
	// flatten to json
	buff, err := json.Marshal(c)
	if err != nil {
		logrus.Errorf("Error while hashing config %s", err)
	}

	m := md5.New()
	m.Write(buff)
	return m.Sum(nil)
}
