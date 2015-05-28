package config

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/eliothedeman/bangarang/alarm"
)

type Configer interface {
	ConfigStruct() interface{}
	Init(interface{}) error
}

var (
	DEFAULT_ENCODING = "json"
)

const (
	DEFAULT_DB_PATH       = "event.db"
	DEFAULT_KEEPALIVE_AGE = "25m"
)

type AppConfig struct {
	EscalationsDir   string                 `json:"escalations_dir"`
	KeepAliveAge     time.Duration          `json:"-"`
	Raw_KeepAliveAge string                 `json:"keep_alive_age"`
	DbPath           string                 `json:"db_path"`
	TcpPort          *int                   `json:"tcp_port"`
	HttpPort         *int                   `json:"http_port"`
	Escalations      *alarm.AlarmCollection `json:"escalations"`
	GlobalPolicy     *alarm.Policy          `json:"global_policy"`
	Encoding         *string                `json:"encoding"`
	Policies         []*alarm.Policy        `json:"-"`
}

func LoadConfigFile(fileName string) (*AppConfig, error) {
	buff, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	ac := &AppConfig{
		Raw_KeepAliveAge: DEFAULT_KEEPALIVE_AGE,
		DbPath:           DEFAULT_DB_PATH,
	}
	ac.Escalations = &alarm.AlarmCollection{}

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

	if ac.Encoding == nil {
		ac.Encoding = &DEFAULT_ENCODING
	}

	if ac.GlobalPolicy != nil {
		ac.GlobalPolicy.Compile()
	}

	return ac, nil
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
