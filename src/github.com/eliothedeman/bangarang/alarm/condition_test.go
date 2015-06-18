package alarm

import (
	"fmt"
	"testing"
	"time"

	"github.com/eliothedeman/bangarang/event"
)

func newTestCondition(g, l, e float64) *Condition {
	c := &Condition{
		Greater: &g,
		Less:    &l,
		Exactly: &e,
	}

	c.init(DEFAULT_GROUP_BY)

	return c
}

func newTestEvent(h, s string, m float64) *event.Event {
	e := &event.Event{
		Host:    h,
		Service: s,
		Metric:  m,
	}
	return e
}

func TestAggregation(t *testing.T) {
	c := newTestCondition(10, -1, 5)
	c.Aggregation = &Aggregation{
		WindowLength: 100,
	}

	c.init(map[string]string{
		"host": `\w+\.(?P<deployment>\w+)\.\w+`,
	})

	for i := 0; i < 110; i++ {
		if c.TrackEvent(newTestEvent(fmt.Sprintf("machine.deployment%d.com", i%10), "service", 1)) {
			t.Error()
		}
	}

	// everything should be at it's limit now, so the next 10 should fail
	for i := 0; i < 10; i++ {
		if !c.TrackEvent(newTestEvent(fmt.Sprintf("machine.deployment%d.com", i%10), "service", 1)) {
			t.Error()
		}
	}
}

func TestAggCloseout(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	c := newTestCondition(10, -1, 5)
	c.Aggregation = &Aggregation{
		WindowLength: 1,
	}

	c.init(map[string]string{
		"host": `\w+\.(?P<deployment>\w+)\.\w+`,
	})

	for i := 0; i < 110; i++ {
		if c.TrackEvent(newTestEvent(fmt.Sprintf("machine.deployment%d.com", i%10), "service", 1)) {
			t.Error()
		}
	}

	// everything should be at it's limit now, so the next 10 should fail
	for i := 0; i < 10; i++ {
		if !c.TrackEvent(newTestEvent(fmt.Sprintf("machine.deployment%d.com", i%10), "service", 1)) {
			t.Error()
		}
	}

	time.Sleep(1 * time.Second)

	// make sure we aren't at the limit
	for i := 0; i < 110; i++ {
		if c.TrackEvent(newTestEvent(fmt.Sprintf("machine.deployment%d.com", i%10), "service", 1)) {
			t.Error()
		}
	}

}

func TestConditionTrackEvent(t *testing.T) {
	c := newTestCondition(100, -100, 5)
	no := newTestEvent("test", "service", 22)
	yes := newTestEvent("test", "service", 1000)

	if c.TrackEvent(no) {
		t.Fail()
	}

	if !c.TrackEvent(yes) {
		t.Fail()
	}

}

func TestGroupingGenName(t *testing.T) {
	g := compileGrouper(DEFAULT_GROUP_BY)

	e := newTestEvent("this", "is", 1)

	last := g.genIndexName(e)

	// insure this name is always consistant
	for i := 0; i < 100; i++ {
		if last != g.genIndexName(e) {
			t.Fail()
		}
	}
}

func TestGroupByHostName(t *testing.T) {
	g := compileGrouper(map[string]string{
		"host": `\w+\.(?P<boom>\w+)\.\w+`,
	})

	e := newTestEvent("my.test.com", "is-fun", 1)
	expected := ":test"

	if g.genIndexName(e) != expected {
		t.Error("expected:", expected, "got:", g.genIndexName(e))
	}
}

func BenchmarkGroupByOne(b *testing.B) {
	g := compileGrouper(map[string]string{
		"host": `\w+\.(?P<boom>\w+)\.\w+`,
	})

	e := newTestEvent("my.test.com", "is-fun", 1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.genIndexName(e)
	}
}

func BenchmarkGroupByTwo(b *testing.B) {
	g := compileGrouper(map[string]string{
		"host":    `\w+\.(?P<boom>\w+)\.\w+`,
		"service": `.*`,
	})

	e := newTestEvent("my.test.com", "is-fun", 1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.genIndexName(e)
	}
}

func BenchmarkGroupByThree(b *testing.B) {
	g := compileGrouper(map[string]string{
		"host":        `\w+\.(?P<boom>\w+)\.\w+`,
		"service":     `.*`,
		"sub_service": `.*`,
	})

	e := newTestEvent("my.test.com", "is-fun", 1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.genIndexName(e)
	}
}
