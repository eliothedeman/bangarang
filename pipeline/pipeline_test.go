package pipeline

import (
	"testing"

	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/alarm/console"
	"github.com/eliothedeman/bangarang/event"
)

func testPipeline(e []*alarm.Escalation) *Pipeline {
	return &Pipeline{
		escalations: e,
	}
}

func testEscalation(crit, warn *alarm.Condition, match, notMatch map[string]string) *alarm.Escalation {
	e := &alarm.Escalation{
		Policy: alarm.Policy{
			Warn:     warn,
			Crit:     crit,
			Match:    match,
			NotMatch: notMatch,
		},
		Alarms: []alarm.Alarm{
			console.NewConsole(),
		},
	}

	e.Policy.Compile()
	return e
}

func testCondition(g, l, e *float64, o int) *alarm.Condition {
	return &alarm.Condition{
		Greater:    g,
		Less:       l,
		Exactly:    e,
		Occurences: o,
	}
}

func test_f(f float64) *float64 {
	return &f
}

func TestOccurences(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 2)
	esc := testEscalation(c, nil, map[string]string{"host": "test"}, nil)
	p := testPipeline([]*alarm.Escalation{esc})

	e := &event.Event{
		Host:    "test",
		Service: "test",
		Metric:  1.0,
	}

	if p.Process(e) != event.OK {
		t.Error("occrences hit too early")
	}

	e = &event.Event{
		Host:    "test",
		Service: "test",
		Metric:  1.0,
	}

	if p.Process(e) != event.CRITICAL {
		t.Error("occrences not hit")
	}

}

func TestProcess(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 0)
	esc := testEscalation(c, nil, map[string]string{"host": "test"}, nil)
	p := testPipeline([]*alarm.Escalation{esc})

	e := &event.Event{
		Host:    "test",
		Service: "test",
		Metric:  1.0,
	}

	if p.Process(e) != event.CRITICAL {
		t.Fail()
	}

}
