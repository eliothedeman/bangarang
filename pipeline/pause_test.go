package pipeline

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/event"
)

func TestPausePipeline(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 1)
	pipe := testPolicy(c, nil, map[string]string{"service": "KeepAlive"}, nil)
	p, _ := testPipeline([]*alarm.Policy{pipe})
	p.Start()

	p.pause()
	p.unpause()

	// make sure we can still insert events
	p.in <- &event.Event{}
}

func TestPausePipelineCache(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	c := testCondition(test_f(0), nil, nil, 1)
	pipe := testPolicy(c, nil, map[string]string{"service": "KeepAlive"}, nil)
	p, _ := testPipeline([]*alarm.Policy{pipe})
	p.Start()
	insert := p.in
	p.pause()
	for i := 0; i < 100; i++ {
		insert <- &event.Event{}
	}

	p.unpause()
}
