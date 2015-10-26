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

func isInSlice(key string, s []string) bool {
	for _, x := range s {
		if x == key {
			return true
		}
	}

	return false
}

func TestTrackerTrackEvent(t *testing.T) {
	e := &event.Event{}
	e.Host = "test"
	e.Service = "service"
	e.SubService = "sub"
	x := NewTracker()
	go x.Start()
	x.TrackEvent(e)

	if !isInSlice(e.Host, x.GetHosts()) {
		t.Fail()
	}
	if !isInSlice(e.Service, x.GetServices()) {
		t.Fail()
	}
}
