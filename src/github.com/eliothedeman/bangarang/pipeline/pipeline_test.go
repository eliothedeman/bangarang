package pipeline

import (
	"fmt"
	"log"
	"runtime"
	"sync"
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

func testPolicy(crit, warn *alarm.Condition, match, notMatch *event.TagSet) *alarm.Policy {

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
	pipe := testPolicy(c, nil, &event.TagSet{{Key: INTERNAL_TAG_NAME, Value: KEEP_ALIVE_INTERNAL_TAG}}, nil)
	p, ta := testPipeline(map[string]*alarm.Policy{"test": pipe})
	defer p.index.Delete()
	e := event.NewEvent()
	e.Tags.Set("host", "one one")
	e.Metric = -1

	p.Pass(e)
	e.Wait()
	// sleep long enough for the keep alives to trip
	time.Sleep(100 * time.Millisecond)

	p.keepAliveAge = time.Millisecond * 15
	p.keepAliveCheckTime = time.Millisecond * 50
	s := make(chan struct{})
	go func() {
		p.checkExpired()
		s <- struct{}{}

	}()

	// wait
	<-s
	time.Sleep(100 * time.Millisecond)

	ta.Do(func(ta *test.TestAlert) {
		if len(ta.Events) != 1 {
			t.Fatal(ta.Events)
		}
	})

}

func TestMatchPolicy(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 1)
	pipe := testPolicy(c, nil, &event.TagSet{{Key: "host", Value: "test"}}, nil)
	p, ta := testPipeline(map[string]*alarm.Policy{"test": pipe})
	defer p.index.Delete()
	e := event.NewEvent()
	e.Tags.Set("host", "test")
	e.Tags.Set("service", "test")
	e.Metric = 1.0

	p.Process(e)
	e.Wait()
	if len(ta.Events) == 0 {
		t.Fatal()
	}
	ta.Do(func(ta *test.TestAlert) {
		for k, _ := range ta.Events {
			if k.IndexName() != e.IndexName() {
				t.Fatal(k.IndexName(), e.IndexName())
			}
		}
	})
}

func TestOccurences(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 2)
	pipe := testPolicy(c, nil, &event.TagSet{{Key: "host", Value: "test"}}, nil)
	p, ta := testPipeline(map[string]*alarm.Policy{"test": pipe})
	defer p.index.Delete()
	e := event.NewEvent()
	e.Tags.Set("host", "test")
	e.Tags.Set("service", "test")
	e.Metric = 1.0

	p.Pass(e)
	e.Wait()

	ta.Do(func(ta *test.TestAlert) {
		if len(ta.Events) != 0 {
			t.Error("occrences hit too early")
		}
	})
	e = event.NewEvent()
	e.Tags.Set("host", "test")
	e.Tags.Set("service", "test")
	e.Metric = 1.0

	p.Pass(e)
	e.Wait()

	ta.Do(func(ta *test.TestAlert) {
		if len(ta.Events) != 1 {
			t.Error("occrences not hit", ta.Events)
		}
	})
}

func genEventSlice(size int) []*event.Event {
	e := make([]*event.Event, size)
	for i := range e {
		e[i] = event.NewEvent()
	}
	return e

}

func BenchmarkProcessOk(b *testing.B) {
	c := testCondition(test_f(10), nil, nil, 0)
	pipe := testPolicy(c, nil, &event.TagSet{{Key: "host", Value: "test"}}, nil)
	p, _ := testPipeline(map[string]*alarm.Policy{"test": pipe})
	defer p.index.Delete()
	e := genEventSlice(b.N)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Pass(e[i])
	}
	for i := 0; i < b.N; i++ {
		e[i].Wait()
	}
}

func BenchmarkProcess2CPU(b *testing.B) {
	c := testCondition(test_f(10), nil, nil, 0)
	pipe := testPolicy(c, nil, &event.TagSet{{Key: "host", Value: "test"}}, nil)
	p, _ := testPipeline(map[string]*alarm.Policy{"test": pipe})
	defer p.index.Delete()
	e := genEventSlice(b.N)

	w := &sync.WaitGroup{}
	w.Add(2)

	b.ReportAllocs()
	b.ResetTimer()
	f := func(s []*event.Event) {
		for i := 0; i < len(s); i++ {
			p.Pass(s[i])
		}
		w.Done()
	}

	go f(e[:len(e)/2])
	go f(e[len(e)/2:])
	w.Wait()
	for i := 0; i < b.N; i++ {
		e[i].Wait()
	}
}

func BenchmarkProcess4CPU(b *testing.B) {
	c := testCondition(test_f(10), nil, nil, 0)
	pipe := testPolicy(c, nil, &event.TagSet{{Key: "host", Value: "test"}}, nil)
	p, _ := testPipeline(map[string]*alarm.Policy{"test": pipe})
	defer p.index.Delete()
	e := genEventSlice(b.N)

	w := &sync.WaitGroup{}
	w.Add(4)

	b.ReportAllocs()
	b.ResetTimer()
	f := func(s []*event.Event) {
		for i := 0; i < len(s); i++ {
			p.Pass(s[i])
		}
		w.Done()
	}

	one := len(e) / 4
	two := one * 2
	three := one + two

	go f(e[:one])
	go f(e[one:two])
	go f(e[two:three])
	go f(e[three:])
	w.Wait()
	for i := 0; i < b.N; i++ {
		e[i].Wait()
	}
}

func BenchmarkIndex(b *testing.B) {
	c := testCondition(test_f(0), nil, nil, 0)
	pipe := testPolicy(c, nil, &event.TagSet{{Key: "host", Value: "test"}}, nil)
	p, _ := testPipeline(map[string]*alarm.Policy{"test": pipe})
	defer p.index.Delete()

	e := event.NewEvent()
	e.Tags.Set("host", "test")
	e.Tags.Set("service", "test")

	e.Metric = -1.0

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Tags.Set("service", fmt.Sprintf("%d", i%1000))
		p.Process(e)
	}

	e.Wait()
}

func TestProcess(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 0)
	pipe := testPolicy(c, nil, &event.TagSet{{Key: "host", Value: "test"}}, nil)
	p, ta := testPipeline(map[string]*alarm.Policy{"test": pipe})
	defer p.index.Delete()
	e := event.NewEvent()
	e.Tags.Set("host", "test")
	e.Tags.Set("service", "test")
	e.Metric = 1.0

	p.Process(e)
	e.Wait()

	if len(ta.Events) != 1 {
		t.Fatal(ta.Events)
	}

	e = event.NewEvent()
	e.Tags.Set("host", "test")
	e.Tags.Set("service", "test")
	e.Metric = -1.0

	p.Process(e)
	e.Wait()
	if ta.Events[e] != event.OK {
		t.Fatal(ta.Events)
	}

}

func TestProcessDedupe(t *testing.T) {
	c := testCondition(test_f(0), nil, nil, 0)
	pipe := testPolicy(c, nil, &event.TagSet{{Key: "host", Value: "test"}}, nil)
	p, ta := testPipeline(map[string]*alarm.Policy{"test": pipe})
	defer p.index.Delete()

	events := make([]*event.Event, 100)

	for i := 0; i < len(events); i++ {
		e := event.NewEvent()
		e.Tags.Set("host", "test")
		e.Tags.Set("service", "test")
		e.Metric = 1.0
		events[i] = e
	}

	p.Process(events[0])

	for i := 1; i < len(events); i++ {
		p.Process(events[i])
		events[i].Wait()
	}

	if len(ta.Events) != 1 {
		log.Println(ta.Events)
		t.Fail()
	}

}
