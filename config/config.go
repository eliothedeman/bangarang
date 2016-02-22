package config

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/escalation"
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
	PutConfig(*AppConfig, *User) (string, error)
	ListSnapshots() []*Snapshot
	GetUser(userName string) (*User, error)
	GetUserByUserName(string) (*User, error)
	DeleteUser(userName string) error
	PutUser(u *User) error
	ListUsers() ([]*User, error)
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
	EscalationsDir  string                                  `json:"escalations_dir"`
	KeepAliveAge    time.Duration                           `json:"-"`
	RawKeepAliveAge string                                  `json:"keep_alive_age"`
	DbPath          string                                  `json:"db_path"`
	Escalations     map[string]*escalation.EscalationPolicy `json:"escalations"`
	Encoding        string                                  `json:"encoding"`
	Policies        map[string]*escalation.Policy           `json:"policies"`
	EventProviders  *provider.EventProviderCollection       `json:"event_providers"`
	LogLevel        string                                  `json:"log_level"`
	APIPort         int                                     `json:"API_port"`
	Hash            []byte                                  `json:"-"`
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
		Escalations:     map[string]*escalation.EscalationPolicy{},
		Policies:        map[string]*escalation.Policy{},
		LogLevel:        defaultLogLevel,
		EventProviders:  &provider.EventProviderCollection{},
	}
}

func loadPolicy(buff []byte) (*escalation.Policy, error) {
	p := &escalation.Policy{}
	err := json.Unmarshal(buff, p)
	if err != nil {
		return p, err
	}

	return p, err
}
