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
	e := event.NewEvent()
	e.Tags.Set("host", "test")
	e.Tags.Set("service", "test")
	e.Tags.Set("sub_service", "test")
	x := NewTracker()
	go x.Start()
	e.WaitInc()
	x.TrackEvent(e)

	if !isInSlice(e.Get("host"), x.GetTag("host")) {
		t.Fail()
	}
	if !isInSlice(e.Get("service"), x.GetTag("service")) {
		t.Fail()
	}
}
