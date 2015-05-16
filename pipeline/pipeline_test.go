package pipeline

import (
	"testing"

	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/event"
)

func TestProcess(t *testing.T) {
	p := newTestPipeline()
	c := &alarm.Condition{}
	a := 0.0
	c.Greater = &a
	p.escalations[0].Policy.Crit = c
	p.escalations[0].Policy.Match["host"] = "test"
	p.escalations[0].Policy.Compile()

	e := &event.Event{
		Host:   "test",
		Metric: 1.0,
	}

	p.Process(e)
	t.Fail()
}
