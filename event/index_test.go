package event

import (
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
	return NewIndex()
}

func newTestEvent(h, s, ss string, m float64) *Event {
	return &Event{
		Host:       h,
		Service:    s,
		SubService: ss,
		Metric:     m,
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
