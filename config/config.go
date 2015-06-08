package config

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/provider"
)

type Configer interface {
	ConfigStruct() interface{}
	Init(interface{}) error
}

var (
	DEFAULT_ENCODING  = "json"
	DEFAULT_LOG_LEVEL = "info"
)

const (
	DEFAULT_DB_PATH       = "event.db"
	DEFAULT_KEEPALIVE_AGE = "25m"
	DEFAULT_API_PORT      = 8081
)

type BasicAuth struct {
	UserName, PasswordHash string
}

type AppConfig struct {
	EscalationsDir   string                            `json:"escalations_dir"`
	KeepAliveAge     time.Duration                     `json:"-"`
	Raw_KeepAliveAge string                            `json:"keep_alive_age"`
	DbPath           string                            `json:"db_path"`
	Escalations      *alarm.AlarmCollection            `json:"escalations"`
	GlobalPolicy     *alarm.Policy                     `json:"global_policy"`
	Encoding         string                            `json:"encoding"`
	Policies         []*alarm.Policy                   `json:"-"`
	EventProviders   *provider.EventProviderCollection `json:"event_providers"`
	LogLevel         string                            `json:"log_level"`
	ApiPort          int                               `json:"api_port"`
	Auths            []BasicAuth                       `json:"basic_auth_users"`
}

func NewDefaultConfig() *AppConfig {
	return &AppConfig{
		Raw_KeepAliveAge: DEFAULT_KEEPALIVE_AGE,
		DbPath:           DEFAULT_DB_PATH,
		ApiPort:          DEFAULT_API_PORT,
		Encoding:         DEFAULT_ENCODING,
		Escalations:      &alarm.AlarmCollection{},
		EventProviders:   &provider.EventProviderCollection{},
	}
}

func parseConfigFile(buff []byte) (*AppConfig, error) {
	var err error
	ac := NewDefaultConfig()

	err = json.Unmarshal(buff, ac)
	if err != nil {
		return nil, err
	}

	ac.KeepAliveAge, err = time.ParseDuration(ac.Raw_KeepAliveAge)
	if err != nil {
		return ac, err
	}

	paths, err := filepath.Glob(ac.EscalationsDir + "*.json")
	if err != nil {
		return nil, err
	}

	for _, path := range paths {
		p, err := loadPolicy(path)
		if err != nil {
			return ac, err
		}

		ac.Policies = append(ac.Policies, p)
	}

	if ac.GlobalPolicy != nil {
		ac.GlobalPolicy.Compile()
	}

	if ac.EventProviders == nil {
		ac.EventProviders = &provider.EventProviderCollection{}
	}

	if ac.LogLevel == "" {
		ac.LogLevel = DEFAULT_LOG_LEVEL
	}

	return ac, nil

}

func LoadConfigFile(fileName string) (*AppConfig, error) {
	buff, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return parseConfigFile(buff)
}

func loadPolicy(fileName string) (*alarm.Policy, error) {
	if !filepath.IsAbs(fileName) {
		fileName, _ = filepath.Abs(fileName)
	}
	buff, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	p := &alarm.Policy{}
	err = json.Unmarshal(buff, p)
	if err != nil {
		return p, err
	}

	if p.Name == "" {
		fileName = filepath.Base(fileName)
		p.Name = fileName[:len(fileName)-4]
	}

	p.Compile()

	return p, err
}
