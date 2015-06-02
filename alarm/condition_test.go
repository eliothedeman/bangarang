package alarm

import (
	"testing"

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
