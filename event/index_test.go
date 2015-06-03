package event

import (
	"fmt"
	"log"
	"testing"
)

var (
	numIndexes = 0
)

func init() {
	log.SetFlags(log.Lshortfile)
}

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

func BenchmarkIndexPutJSON(b *testing.B) {
	index := newTestIndex()
	index.pool = NewEncodingPool(NewJsonEncoder, NewJsonDecoder, 4)
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

func BenchmarkIndexPutMSGP(b *testing.B) {
	index := newTestIndex()
	index.pool = NewEncodingPool(NewMsgPackEncoder, NewMsgPackDecoder, 4)
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
func TestDeleteIncidentById(t *testing.T) {
	i := newTestIndex()
	defer i.Delete()
	e := newTestEvent("h", "s", "ss", 1)
	in := NewIncident("DeleteIncident", e)
	i.PutIncident(in)
	in = i.GetIncident(in.IndexName())
	if in == nil {
		t.Fail()
	}

	i.DeleteIncidentById(in.IndexName())

	in = i.GetIncident(in.IndexName())
	if in != nil {
		t.Fail()
	}
}

func TestListIncidents(t *testing.T) {
	i := newTestIndex()
	defer i.Delete()
	e := newTestEvent("h", "s", "ss", 1)
	in := NewIncident("ListIncidents", e)
	i.PutIncident(in)

	ins := i.ListIncidents()

	if len(ins[0].EventName) != len(in.EventName) {
		t.Fail()
	}
}

func TestAddIncident(t *testing.T) {
	i := newTestIndex()
	defer i.Delete()
	e := newTestEvent("h", "s", "ss", 1)
	in := NewIncident("test", e)
	i.PutIncident(in)

	b := i.GetIncident(in.IndexName())

	if len(in.EventName) != len(b.EventName) {
		t.Fail()
	}
}
