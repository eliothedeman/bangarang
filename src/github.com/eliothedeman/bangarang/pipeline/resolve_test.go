package pipeline

import (
	"testing"

	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/event"
)

func TestResolve(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 1)
	pipe := testPolicy(c, nil, map[string]string{"host": "test"}, nil)
	p, ta := testPipeline([]*alarm.Policy{pipe})
	defer p.index.Delete()

	e := &event.Event{}
	e.Metric = 1
	e.Host = "test"

	p.Process(e)
	if len(ta.Events) != 1 {
		t.Fail()
	}

	ta.Events = make(map[*event.Event]int)

	e.Metric = 0
	p.Process(e)

	if len(ta.Events) != 1 {
		t.Fail()
	}

}
