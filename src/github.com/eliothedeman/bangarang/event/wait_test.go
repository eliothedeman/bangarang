package event

import (
	"testing"
	"time"
)

func TestGetState(t *testing.T) {
	e := NewEvent()

	if e.GetState() != StateStart {
		t.Fail()
	}
}

func TestSetState(t *testing.T) {
	e := NewEvent()
	if e.GetState() != StateStart {
		t.Fail()
	}

	e.SetState(StateComplete)

	if e.GetState() != StateComplete {
		t.Fail()
	}
}

func TestWaitForState(t *testing.T) {
	e := NewEvent()

	f := e.WaitForState(StatePolicy, 20*time.Millisecond)

	// make sure the timeout will return
	f()

	f = e.WaitForState(StateComplete, time.Second)

	go func() {
		e.SetState(StateComplete)
	}()

	f()
}
