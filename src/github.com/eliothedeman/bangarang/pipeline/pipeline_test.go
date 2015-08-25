package pipeline

import (
	"fmt"
	"log"
	"runtime"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/alarm/test"
	"github.com/eliothedeman/bangarang/event"
)

var (
	tests_ran = 100
)

func testPipeline(p map[string]*alarm.Policy) (*Pipeline, *test.TestAlert) {
	tests_ran += 1
	ta := test.NewTest().(*test.TestAlert)
	pipe := &Pipeline{
		policies:     p,
		index:        event.NewIndex(),
		pauseChan:    make(chan struct{}),
		unpauseChan:  make(chan struct{}),
		encodingPool: event.NewEncodingPool(event.EncoderFactories["json"], event.DecoderFactories["json"], runtime.NumCPU()),
		escalations: &alarm.Collection{
			Coll: map[string][]alarm.Alarm{
				"test": []alarm.Alarm{ta},
			},
		},
		tracker: NewTracker(),
		in:      make(chan *event.Event),
	}

	go pipe.tracker.Start()
	pipe.Start()
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

func TestKeepAlive(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	c := testCondition(test_f(0), nil, nil, 1)
	pipe := testPolicy(c, nil, map[string]string{"service": "KeepAlive"}, nil)
	p, ta := testPipeline(map[string]*alarm.Policy{"test": pipe})
	defer p.index.Delete()
	e := event.NewEvent()
	e.Host = "one one"
	e.Service = "exit"
	e.Metric = -1

	p.Pass(e)
	e.Wait()

	p.keepAliveAge = time.Millisecond * 15
	p.keepAliveCheckTime = time.Millisecond * 50
	go p.checkExpired()
	time.Sleep(100 * time.Millisecond)

	if len(ta.Events) != 1 {
		t.Fatal(ta.Events)
	}

}

func TestMatchPolicy(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 1)
	pipe := testPolicy(c, nil, map[string]string{"host": "test"}, nil)
	p, ta := testPipeline(map[string]*alarm.Policy{"test": pipe})
	defer p.index.Delete()
	e := event.NewEvent()
	e.Host = "test"
	e.Service = "test"
	e.Metric = 1.0

	p.Process(e)
	e.Wait()
	if len(ta.Events) == 0 {
		t.Fatal()
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
	p, ta := testPipeline(map[string]*alarm.Policy{"test": pipe})
	defer p.index.Delete()
	e := event.NewEvent()
	e.Host = "test"
	e.Service = "test"
	e.Metric = 1.0

	p.Process(e)
	e.Wait()

	if len(ta.Events) != 0 {
		t.Error("occrences hit too early")
	}
	e = event.NewEvent()
	e.Host = "test"
	e.Service = "test"
	e.Metric = 1.0

	e.Wait()

	if len(ta.Events) != 1 {
		t.Error("occrences not hit")
	}
}

func BenchmarkProcessOk(b *testing.B) {
	c := testCondition(test_f(0), nil, nil, 0)
	pipe := testPolicy(c, nil, map[string]string{"host": "test"}, nil)
	p, _ := testPipeline(map[string]*alarm.Policy{"test": pipe})
	defer p.index.Delete()

	e := event.NewEvent()
	e.Host = "test"
	e.Service = "test"
	e.Metric = -1.0

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Process(e)
	}
	e.Wait()
}

func BenchmarkIndex(b *testing.B) {
	c := testCondition(test_f(0), nil, nil, 0)
	pipe := testPolicy(c, nil, map[string]string{"host": "test"}, nil)
	p, _ := testPipeline(map[string]*alarm.Policy{"test": pipe})
	defer p.index.Delete()

	e := event.NewEvent()

	e.Host = "test"
	e.Service = "test"
	e.Metric = -1.0

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Service = fmt.Sprintf("%d", i%1000)
		p.Process(e)
	}

	e.Wait()
}

func TestProcess(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 0)
	pipe := testPolicy(c, nil, map[string]string{"host": "test"}, nil)
	p, ta := testPipeline(map[string]*alarm.Policy{"test": pipe})
	defer p.index.Delete()
	e := event.NewEvent()
	e.Host = "test"
	e.Service = "test"
	e.Metric = 1.0

	p.Process(e)
	e.Wait()

	if len(ta.Events) != 1 {
		t.Fail()
	}

	e = event.NewEvent()
	e.Host = "test"
	e.Service = "test"
	e.Metric = -1.0

	p.Process(e)
	e.Wait()

	if len(ta.Events) != 0 {
		t.Fail()
	}

}

func TestProcessDedupe(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 0)
	pipe := testPolicy(c, nil, map[string]string{"host": "test"}, nil)
	p, ta := testPipeline(map[string]*alarm.Policy{"test": pipe})
	defer p.index.Delete()

	events := make([]*event.Event, 100)

	for i := 0; i < len(events); i++ {
		e := event.NewEvent()
		e.Host = "test"
		e.Service = "test"
		e.Metric = 1.0
		events[i] = e
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
