package event

import (
	"fmt"
	"testing"
)

var (
	numIndexes = 0
)

func newTestIndex() *Index {
	numIndexes += 1
	return NewIndex(fmt.Sprintf("test%d.db", numIndexes))
}

func newTestEvent(h, s, ss string, m float64) *Event {
	return &Event{
		Host:       h,
		Service:    s,
		SubService: ss,
		Metric:     m,
	}
}

func TestIndexPut(t *testing.T) {
	i := newTestIndex()
	defer i.Delete()

	e1 := newTestEvent("h", "s", "ss", 0.0)
	e2 := newTestEvent("h", "s", "ss", 0.0)

	i.PutEvent(e1)
	i.PutEvent(e2)

	e3 := i.GetEvent([]byte(e2.IndexName()))

	if e3.LastEvent == nil {
		t.Fail()
	}

}

func BenchmarkIndexPut(b *testing.B) {
	index := newTestIndex()
	defer index.Delete()
	events := make([]*Event, 1000)

	for i := 0; i < 1000; i++ {
		events[i] = newTestEvent("h", "s", "ss", float64(i))
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		index.PutEvent(events[i%1000])
	}
}

func TestListIncidents(t *testing.T) {
	i := newTestIndex()
	defer i.Delete()
	e := newTestEvent("h", "s", "ss", 1)
	in := NewIncident("ListIncidents", i, e)

	ins := i.ListIncidents()

	if ins[0].EventName != in.EventName {
		t.Fail()
	}

}

func TestAddIncident(t *testing.T) {
	i := newTestIndex()
	defer i.Delete()
	e := newTestEvent("h", "s", "ss", 1)
	in := NewIncident("test", i, e)

	b := i.GetIncident(in.Id)

	if in.EventName != b.EventName {
		t.Fail()
	}

}
