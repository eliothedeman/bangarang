package pipeline

import (
	"testing"

	"github.com/eliothedeman/bangarang/escalation"
	"github.com/eliothedeman/bangarang/event"
)

func TestResolve(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 1)
	pipe := testPolicy(c, nil, &event.TagSet{{Key: "host", Value: "test"}}, nil)
	p, ta := testPipeline(map[string]*escalation.Policy{"test": pipe})
	defer p.index.Delete()

	e := event.NewEvent()
	e.Tags.Set("host", "test")
	e.Metric = 1

	p.processEvent(e)
	e.Wait()
	if len(ta.Events) != 1 {
		t.Fatal(ta.Events)
	}

	ta.Events = make(map[*event.Event]int)

	e.Metric = 0
	p.processEvent(e)
	e.Wait()

	if len(ta.Events) != 1 {
		t.Fatal(ta.Events)
	}

}
