package config

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/eliothedeman/bangarang/alarm"
)

type Configer interface {
	ConfigStruct() interface{}
	Init(interface{}) error
}

type AppConfig struct {
	Escalations    []*alarm.Escalation    `json:"-"`
	EscalationsDir string                 `json:"escalations_dir"`
	TcpPort        *int                   `json:"tcp_port"`
	HttpPort       *int                   `json:"http_port"`
	Alarms         *alarm.AlarmCollection `json:"alarms"`
}

func LoadConfigFile(fileName string) (*AppConfig, error) {
	buff, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	ac := &AppConfig{}
	ac.Alarms = &alarm.AlarmCollection{}

	err = json.Unmarshal(buff, ac)
	if err != nil {
		return nil, err
	}

	paths, err := filepath.Glob(ac.EscalationsDir + "*.json")
	if err != nil {
		return nil, err
	}

	for _, path := range paths {
		e, err := loadEscalation(path)
		if err != nil {
			return ac, err
		}

		ac.Escalations = append(ac.Escalations, e)
	}

	return ac, nil
}

func loadEscalation(fileName string) (*alarm.Escalation, error) {
	if !filepath.IsAbs(fileName) {
		fileName, _ = filepath.Abs(fileName)
	}
	buff, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	e := &alarm.Escalation{}
	err = json.Unmarshal(buff, e)
	if err != nil {
		return e, err
	}
	e.Policy.Compile()

	err = e.LoadAlarms()
	return e, err
}
