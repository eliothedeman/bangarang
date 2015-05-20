package event

import "testing"

func TestEventDedup(t *testing.T) {
	e1 := newTestEvent("h", "s", "ss", 0)
	e1.Status = CRITICAL
	e2 := newTestEvent("h", "s", "ss", 0)
	e2.Status = CRITICAL
	e1.LastEvent = e2

	if e1.StatusChanged() {
		t.Fail()
	}

}
