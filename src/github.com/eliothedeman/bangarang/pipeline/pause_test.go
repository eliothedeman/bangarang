package pipeline

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/escalation"
	_ "github.com/eliothedeman/bangarang/escalation/console"
	"github.com/eliothedeman/bangarang/event"
)

func TestPausePipeline(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 1)
	pipe := testPolicy(c, nil, &event.TagSet{{Key: "service", Value: "KeepAlive"}}, nil)
	p, _ := testPipeline(map[string]*escalation.Policy{"test": pipe})
	defer p.index.Delete()
	p.Pause()
	p.Unpause()

	// make sure we can still insert events
	p.PassEvent(event.NewEvent())
}

func TestAddConfig(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 1)
	pipe := testPolicy(c, nil, &event.TagSet{{Key: "service", Value: "KeepAlive"}}, nil)
	p, _ := testPipeline(map[string]*escalation.Policy{"test": pipe})
	defer p.index.Delete()
	p.processEvent(event.NewEvent())

	conf := &config.AppConfig{}
	conf.Policies = map[string]*escalation.Policy{"new": testPolicy(c, nil, &event.TagSet{{Key: "2", Value: "2"}}, nil)}

	p.Refresh(conf)
	p.PassEvent(event.NewEvent())

	p.ViewConfig(func(ac *config.AppConfig) {
		conf = ac
	})

	log.Println(p.policies)
	conf.Policies["other"] = testPolicy(c, nil, &event.TagSet{{Key: "1", Value: "1"}}, nil)
	p.Refresh(conf)
	log.Println(p.policies)
	for i := 0; i < 100; i++ {
		p.PassEvent(&event.Event{})

	}

}

func TestPausePipelineCache(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	c := testCondition(test_f(0), nil, nil, 1)
	pipe := testPolicy(c, nil, &event.TagSet{{Key: "service", Value: "wwoo"}}, nil)
	p, _ := testPipeline(map[string]*escalation.Policy{"test": pipe})
	defer p.index.Delete()
	p.Start()
	p.Pause()
	for i := 0; i < 100; i++ {
		p.PassEvent(event.NewEvent())
	}

	p.Unpause()
}

func TestRefreshPipeline(t *testing.T) {
	one := []byte(`{
	"api_port": 8082,
	"escalations": {
		"testing": {
			"configs": [
				{
					"type": "console"
				}
			]
		}
	},
	"keep_alive_age": "10s",
    "escalations_dir": "alerts/"
}`)
	ac := config.NewDefaultConfig()

	err := json.Unmarshal(one, ac)
	if err != nil {
		t.Error(err)
	}

	p := NewPipeline(ac)
	defer p.index.Delete()
	ac.Policies = map[string]*escalation.Policy{"test": &escalation.Policy{}}

	p.Refresh(ac)

}
