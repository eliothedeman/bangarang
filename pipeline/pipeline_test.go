package pipeline

import (
	"fmt"
	"log"
	"runtime"
	"testing"

	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/alarm/test"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/event"
)

var (
	tests_ran = 0
)

func testPipeline(p []*alarm.Policy) (*Pipeline, *test.TestAlert) {
	tests_ran += 1
	ta := test.NewTest().(*test.TestAlert)
	pipe := &Pipeline{
		policies:     p,
		index:        event.NewIndex(fmt.Sprintf("test%d.db", tests_ran)),
		encodingPool: event.NewEncodingPool(event.EncoderFactories[config.DEFAULT_ENCODING], event.DecoderFactories[config.DEFAULT_ENCODING], runtime.NumCPU()),
		escalations: map[string][]alarm.Alarm{
			"test": []alarm.Alarm{ta},
		},
	}
	return pipe, ta
}

func testPolicy(crit, warn *alarm.Condition, match, notMatch map[string]string) *alarm.Policy {
	p := &alarm.Policy{
		Warn:     warn,
		Crit:     crit,
		Match:    match,
		NotMatch: notMatch,
	}

	p.Compile()
	return p
}

func testCondition(g, l, e *float64, o int) *alarm.Condition {
	return &alarm.Condition{
		Greater:    g,
		Less:       l,
		Exactly:    e,
		Occurences: o,
		Escalation: "test",
	}
}

func test_f(f float64) *float64 {
	return &f
}

func TestMatchPolicy(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 1)
	pipe := testPolicy(c, nil, map[string]string{"host": "test"}, nil)
	p, ta := testPipeline([]*alarm.Policy{pipe})
	defer p.index.Delete()

	e := &event.Event{
		Host:    "test",
		Service: "test",
		Metric:  1.0,
	}

	p.Process(e)
	if len(ta.Events) == 0 {
		t.Fail()
	}
	for k, _ := range ta.Events {
		if k.IndexName() != e.IndexName() {
			t.Fail()
		}
	}
}

func TestOccurences(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 2)
	pipe := testPolicy(c, nil, map[string]string{"host": "test"}, nil)
	p, _ := testPipeline([]*alarm.Policy{pipe})
	defer p.index.Delete()

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

func BenchmarkProcessOk(b *testing.B) {
	c := testCondition(test_f(0), nil, nil, 0)
	pipe := testPolicy(c, nil, map[string]string{"host": "test"}, nil)
	p, _ := testPipeline([]*alarm.Policy{pipe})
	defer p.index.Delete()

	e := &event.Event{
		Host:    "test",
		Service: "test",
		Metric:  -1.0,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Process(e)
	}
}

func BenchmarkIndex(b *testing.B) {
	c := testCondition(test_f(0), nil, nil, 0)
	pipe := testPolicy(c, nil, map[string]string{"host": "test"}, nil)
	p, _ := testPipeline([]*alarm.Policy{pipe})
	defer p.index.Delete()

	e := &event.Event{
		Host:    "test",
		Service: "test",
		Metric:  -1.0,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Service = fmt.Sprintf("%d", i)
		p.Process(e)
	}

}

func BenchmarkIndexWithStats(b *testing.B) {
	c := testCondition(test_f(0), nil, nil, 0)
	c.StdDev = &alarm.StdDev{
		Sigma: 4,
	}

	pipe := testPolicy(c, nil, map[string]string{"host": "test"}, nil)
	p, _ := testPipeline([]*alarm.Policy{pipe})
	defer p.index.Delete()

	e := &event.Event{
		Host:    "test",
		Service: "test",
		Metric:  -1.0,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Service = fmt.Sprintf("%d", i)
		p.Process(e)
	}

}

func TestProcess(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 0)
	pipe := testPolicy(c, nil, map[string]string{"host": "test"}, nil)
	p, _ := testPipeline([]*alarm.Policy{pipe})
	defer p.index.Delete()

	e := &event.Event{
		Host:    "test",
		Service: "test",
		Metric:  1.0,
	}

	if p.Process(e) != event.CRITICAL {
		t.Fail()
	}

	e = &event.Event{
		Host:    "testok",
		Service: "testok",
		Metric:  -1.0,
	}

	if p.Process(e) != event.OK {
		t.Fail()
	}

}

func TestProcessDedupe(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 0)
	pipe := testPolicy(c, nil, map[string]string{"host": "test"}, nil)
	p, ta := testPipeline([]*alarm.Policy{pipe})
	defer p.index.Delete()

	events := make([]*event.Event, 100)

	for i := 0; i < len(events); i++ {
		events[i] = &event.Event{
			Host:   "test",
			Metric: 1.0,
		}
	}

	p.Process(events[0])

	for i := 1; i < len(events); i++ {
		p.Process(events[i])
	}

	if len(ta.Events) != 1 {
		log.Println(ta.Events)
		t.Fail()
	}

}
