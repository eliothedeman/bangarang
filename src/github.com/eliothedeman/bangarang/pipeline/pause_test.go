package pipeline

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/alarm"
	_ "github.com/eliothedeman/bangarang/alarm/console"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/event"
)

func TestPausePipeline(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 1)
	pipe := testPolicy(c, nil, map[string]string{"service": "KeepAlive"}, nil)
	p, _ := testPipeline(map[string]*alarm.Policy{"test": pipe})
	defer p.index.Delete()
	p.Pause()
	p.Unpause()

	// make sure we can still insert events
	p.Pass(&event.Event{})
}

func TestPausePipelineCache(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	c := testCondition(test_f(0), nil, nil, 1)
	pipe := testPolicy(c, nil, map[string]string{"service": "wwoo"}, nil)
	p, _ := testPipeline(map[string]*alarm.Policy{"test": pipe})
	defer p.index.Delete()
	p.Start()
	p.Pause()
	for i := 0; i < 100; i++ {
		p.Pass(event.NewEvent())
	}

	p.Unpause()
}

func TestRefreshPipeline(t *testing.T) {
	one := []byte(`{
	"api_port": 8082,
	"escalations": {
		"testing": [
			{
				"type": "console"
			}
		]
	},
	"keep_alive_age": "10s",
    "escalations_dir": "alerts/"
}`)
	ac, err := config.ParseConfigFile(one)
	if err != nil {
		t.Error(err)
	}

	p := NewPipeline(ac)
	defer p.index.Delete()
	ac.Policies = map[string]*alarm.Policy{"test": &alarm.Policy{}}

	p.Refresh(ac)

}
