package pipeline

import (
	"testing"

	"github.com/eliothedeman/bangarang/event"
)

func TestTrackerStart(t *testing.T) {
	x := NewTracker()
	if x == nil {
		t.Fail()
	}
}

func TestTrackerTrackEvent(t *testing.T) {
	e := &event.Event{}
	e.Host = "test"
	e.Service = "service"
	e.SubService = "sub"
	x := NewTracker()
	go x.Start()
	x.TrackEvent(e)

	if x.hosts[e.Host].get() != 1 {
		t.Fail()
	}
	if x.services[e.Service].get() != 1 {
		t.Fail()
	}
	if x.subServices[e.SubService].get() != 1 {
		t.Fail()
	}

}
