package config

import (
	"crypto/md5"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
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
	Hash             []byte                            `json:"-"`
	fileName         string
}

func (c *AppConfig) FileName() string {
	return c.fileName
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

	ac.KeepAliveAge, err = time.ParseDuration(ac.Raw_KeepAliveAge)
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
