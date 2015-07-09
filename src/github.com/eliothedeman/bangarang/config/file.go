package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
)

type FileConf struct {
	path string
}

func (f *FileConf) versionPath(version string) string {
	return fmt.Sprintf("%s/%s.json", f.path, version)
}

func (f *FileConf) getRawFile(version string) ([]byte, error) {
	return ioutil.ReadFile(f.versionPath(version))
}

func (f *FileConf) encode(i interface{}) ([]byte, error) {
	return json.Marshal(i)
}

func (f *FileConf) decode(buff []byte, i interface{}) error {
	return json.Unmarshal(buff, i)
}

func (f *FileConf) GetCurrent() (*AppConfig, error) {
	a, err := f.GetVersion("current")
	if err != nil {
		logrus.Error(err)
		a = NewDefaultConfig()
		for _, p := range a.Policies {
			p.Compile()
		}
		a.provider = f
	}

	return a, nil
}

func (f *FileConf) initPath() error {
	return os.MkdirAll(f.path, 0775)
}

func (f *FileConf) PutConfig(ac *AppConfig) (string, error) {
	// get the current config
	c, err := f.GetCurrent()
	if err != nil {
		return "", err
	}

	s := newSnapshot(c)
	buff, err := f.encode(s)
	if err != nil {
		logrus.Error(err)
		return "", err
	}

	err = ioutil.WriteFile(f.versionPath(s.Hash), buff, 0775)
	if err != nil {
		logrus.Error(err)
		return "", err
	}

	// put the current config at "current"
	s = newSnapshot(ac)
	buff, err = f.encode(s)
	if err != nil {
		logrus.Error(err)
		return "", err
	}

	err = ioutil.WriteFile(f.versionPath("current"), buff, 0775)
	return s.Hash, err
}

func (f *FileConf) ListSnapshots() []*Snapshot {
	snaps := []*Snapshot{}
	paths, err := filepath.Glob(f.versionPath("*"))
	if err != nil {
		logrus.Error(err)
		return snaps
	}

	for _, p := range paths {
		p, _ = filepath.Abs(p)
		buff, err := ioutil.ReadFile(p)
		if err != nil {
			logrus.Error(err)
			return snaps
		}
		s := newSnapshot(NewDefaultConfig())
		err = f.decode(buff, s)
		if err != nil {
			logrus.Error(err)
			return snaps
		}

		if strings.Contains(p, "current") {
			s.Hash = "current"
		}

		snaps = append(snaps, s)
	}

	return snaps
}

func (f *FileConf) GetConfig(version string) (*AppConfig, error) {
	return f.GetVersion(version)
}

func (f *FileConf) GetVersion(version string) (*AppConfig, error) {
	buff, err := f.getRawFile(version)
	if err != nil {
		return nil, err
	}

	s := newSnapshot(NewDefaultConfig())
	if len(buff) == 0 {
		return s.App, nil
	}

	err = f.decode(buff, s)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for _, p := range s.App.Policies {
		p.Compile()
	}

	s.App.provider = f

	return s.App, err
}
