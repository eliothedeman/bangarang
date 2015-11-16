package escalation

import (
	"fmt"
	"testing"
	"time"

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

func TestAggregation(t *testing.T) {
	c := newTestCondition(10, -1, 500)
	c.Aggregation = &Aggregation{
		WindowLength: 110,
	}

	c.init(&event.TagSet{
		{"host", `\w+\.(?P<deployment>\w+)\.\w+`},
	})

	for i := 0; i < 110; i++ {
		if c.TrackEvent(newTestEvent(fmt.Sprintf("machine.deployment%d.com", i%10), "service", 1)) {
			t.Fatal(i)
		}
	}

	// everything should be at it's limit now, so the next 10 should fail
	for i := 0; i < 10; i++ {
		if !c.TrackEvent(newTestEvent(fmt.Sprintf("machine.deployment%d.com", i%10), "service", 1)) {
			t.Error()
		}
	}
}

func TestAggCloseout(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	c := newTestCondition(10, -1, 500)
	c.Aggregation = &Aggregation{
		WindowLength: 1,
	}

	c.init(&event.TagSet{
		{"host", `\w+\.(?P<deployment>\w+)\.\w+`},
	})

	for i := 0; i < 110; i++ {
		if c.TrackEvent(newTestEvent(fmt.Sprintf("machine.deployment%d.com", i%10), "service", 1)) {
			t.Fatal()
		}
	}

	// everything should be at it's limit now, so the next 10 should fail
	for i := 0; i < 10; i++ {
		if !c.TrackEvent(newTestEvent(fmt.Sprintf("machine.deployment%d.com", i%10), "service", 1)) {
			t.Error()
		}
	}

	time.Sleep(1 * time.Second)

	// make sure we aren't at the limit
	for i := 0; i < 110; i++ {
		if c.TrackEvent(newTestEvent(fmt.Sprintf("machine.deployment%d.com", i%10), "service", 1)) {
			t.Error()
		}
	}

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
