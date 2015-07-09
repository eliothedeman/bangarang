package config

import (
	"crypto/md5"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/provider"
)

// A Configer provides an interface to dynamicaly load configuration
type Configer interface {
	ConfigStruct() interface{}
	Init(interface{}) error
}

// A Provider can read and write configs, along with querying for versions
type Provider interface {
	GetConfig(version string) (*AppConfig, error)
	GetCurrent() (*AppConfig, error)
	PutConfig(*AppConfig) (string, error)
	ListSnapshots() []*Snapshot
}

// GetProvider returns a config provider at that given path.
// If nothing exists at that path, one will be created.
// If a bad kind id provided, a nil value will be returned
func GetProvider(kind string, path string) Provider {
	switch strings.ToLower(kind) {
	case "db":
		d := &DBConf{}
		d.fileName = path
		err := d.initDB()
		if err != nil {
			logrus.Error(err)
			return nil
		}

		return d

	case "json":
		f := &FileConf{}
		f.path = path
		err := f.initPath()
		if err != nil {
			logrus.Error(err)
			return nil
		}
		return f
	}

	logrus.Errorf("Unknown config provider type %s", kind)
	return nil
}

var (
	defaultEncoding = "json"
	defaultLogLevel = "info"
)

const (
	defaultDBPath       = "event.db"
	defaultKeepaliveAge = "25m"
	defaultAPIPort      = 8081
)

// AppConfig provides configuration options for setting up the application
type AppConfig struct {
	EscalationsDir  string                            `json:"escalations_dir"`
	KeepAliveAge    time.Duration                     `json:"-"`
	RawKeepAliveAge string                            `json:"keep_alive_age"`
	DbPath          string                            `json:"db_path"`
	Escalations     *alarm.Collection                 `json:"escalations"`
	GlobalPolicy    *alarm.Policy                     `json:"global_policy"`
	Encoding        string                            `json:"encoding"`
	Policies        map[string]*alarm.Policy          `json:"policies"`
	EventProviders  *provider.EventProviderCollection `json:"event_providers"`
	LogLevel        string                            `json:"log_level"`
	APIPort         int                               `json:"API_port"`
	Hash            []byte                            `json:"-"`
	fileName        string
	provider        Provider
}

// SetProvider changes the AppConfigs provider to the givin one
func (a *AppConfig) SetProvider(p Provider) {
	a.provider = p
}

// Provider returns the Provider that created this AppConfig
func (c *AppConfig) Provider() Provider {
	return c.provider
}

// FileName returns the name of the file that was used to create this AppConfig
func (c *AppConfig) FileName() string {
	return c.fileName
}

// NewDefaultConfig create and return a new instance of the default configuration
func NewDefaultConfig() *AppConfig {
	return &AppConfig{
		RawKeepAliveAge: defaultKeepaliveAge,
		DbPath:          defaultDBPath,
		APIPort:         defaultAPIPort,
		Encoding:        defaultEncoding,
		Escalations:     &alarm.Collection{},
		LogLevel:        defaultLogLevel,
		EventProviders:  &provider.EventProviderCollection{},
	}
}

func ParseConfigFile(buff []byte) (*AppConfig, error) {
	var err error
	ac := NewDefaultConfig()

	// this will be used to hash all the files thar are opened while parsing
	hasher := md5.New()
	hasher.Write(buff)

	err = json.Unmarshal(buff, ac)
	if err != nil {
		return nil, err
	}

	ac.KeepAliveAge, err = time.ParseDuration(ac.RawKeepAliveAge)
	if err != nil {
		return ac, err
	}

	paths, err := filepath.Glob(ac.EscalationsDir + "*.json")
	if err != nil {
		return nil, err
	}

	for _, path := range paths {
		buff, err := loadFile(path)
		if err != nil {
			return ac, err
		}

		hasher.Write(buff)
		p, err := loadPolicy(buff)
		if err != nil {
			return ac, err
		}

		// set up the file name for the policy
		if p.Name == "" {
			path = filepath.Base(path)
			p.Name = path[:len(path)-4]
		}

		ac.Policies[p.Name] = p
	}

	if ac.GlobalPolicy != nil {
		ac.GlobalPolicy.Compile()
	}

	if ac.EventProviders == nil {
		ac.EventProviders = &provider.EventProviderCollection{}
	}

	if ac.LogLevel == "" {
		ac.LogLevel = defaultLogLevel
	}

	ac.Hash = hasher.Sum(nil)

	return ac, nil

}

func LoadConfigFile(fileName string) (*AppConfig, error) {
	buff, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	ac, err := ParseConfigFile(buff)
	if err != nil {
		logrus.Error(err)
		return ac, err
	}

	ac.fileName = fileName
	return ac, err

}

func loadFile(fileName string) ([]byte, error) {
	if !filepath.IsAbs(fileName) {
		fileName, _ = filepath.Abs(fileName)

	}
	return ioutil.ReadFile(fileName)
}

func loadPolicy(buff []byte) (*alarm.Policy, error) {
	p := &alarm.Policy{}
	err := json.Unmarshal(buff, p)
	if err != nil {
		return p, err
	}

	p.Compile()

	return p, err
}
