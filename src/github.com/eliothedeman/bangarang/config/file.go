package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

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
	return os.MkdirAll(f.path, 0660)
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

//
// func ParseConfigFile(buff []byte) (*AppConfig, error) {
// 	var err error
// 	ac := NewDefaultConfig()
//
// 	// this will be used to hash all the files thar are opened while parsing
// 	hasher := md5.New()
// 	hasher.Write(buff)
//
// 	err = json.Unmarshal(buff, ac)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	ac.KeepAliveAge, err = time.ParseDuration(ac.Raw_KeepAliveAge)
// 	if err != nil {
// 		return ac, err
// 	}
//
// 	paths, err := filepath.Glob(ac.EscalationsDir + "*.json")
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	for _, path := range paths {
// 		buff, err := loadFile(path)
// 		if err != nil {
// 			return ac, err
// 		}
//
// 		hasher.Write(buff)
// 		p, err := loadPolicy(buff)
// 		if err != nil {
// 			return ac, err
// 		}
//
// 		// set up the file name for the policy
// 		if p.Name == "" {
// 			path = filepath.Base(path)
// 			p.Name = path[:len(path)-4]
// 		}
//
// 		ac.Policies = append(ac.Policies, p)
// 	}
//
// 	if ac.GlobalPolicy != nil {
// 		ac.GlobalPolicy.Compile()
// 	}
//
// 	if ac.EventProviders == nil {
// 		ac.EventProviders = &provider.EventProviderCollection{}
// 	}
//
// 	if ac.LogLevel == "" {
// 		ac.LogLevel = DefaultLogLevel
// 	}
//
// 	ac.Hash = hasher.Sum(nil)
//
// 	return ac, nil
//
// }
//
// func LoadConfigFile(fileName string) (*AppConfig, error) {
// 	buff, err := ioutil.ReadFile(fileName)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	ac, err := ParseConfigFile(buff)
// 	if err != nil {
// 		logrus.Error(err)
// 		return ac, err
// 	}
//
// 	ac.fileName = fileName
// 	return ac, err
//
// }
//
// func loadFile(fileName string) ([]byte, error) {
// 	if !filepath.IsAbs(fileName) {
// 		fileName, _ = filepath.Abs(fileName)
//
// 	}
// 	return ioutil.ReadFile(fileName)
// }
