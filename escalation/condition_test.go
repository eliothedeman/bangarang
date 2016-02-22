package escalation

import (
	"testing"

	"github.com/eliothedeman/bangarang/event"
)

func newTestCondition(g, l, e float64) *Condition {
	c := &Condition{
		Greater:    &g,
		Less:       &l,
		Exactly:    &e,
		WindowSize: 100,
	}

	c.init(DEFAULT_GROUP_BY)

	return c
}

func newTestEvent(h, s string, m float64) *event.Event {
	e := event.NewEvent()
	e.Tags.Set("host", h)
	e.Tags.Set("service", s)
	e.Metric = m
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

func TestNonNegativeDerivative(t *testing.T) {
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
