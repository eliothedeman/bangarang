package escalation

import (
	"log"
	"math/rand"
	"testing"
)

func TestStdDevFalsePositive(t *testing.T) {
	c := &Condition{
		Greater:    test_f(5),
		StdDev:     true,
		WindowSize: 100,
	}

	c.init(DEFAULT_GROUP_BY)
	for i := 0; i < 1000; i++ {
		e := newTestEvent("machine.test.com", "test_service", rand.Float64()*10)
		if c.TrackEvent(e) {
			log.Println(i)
			c.DoOnTracker(e, func(et *eventTracker) {
				t.Fatal(et.df.Data())
			})
		}
	}
}

func TestStdDev(t *testing.T) {
	c := &Condition{
		Greater:    test_f(5),
		StdDev:     true,
		WindowSize: 100,
	}

	c.init(DEFAULT_GROUP_BY)

	for i := 0; i < 100; i++ {
		e := newTestEvent("machine.test.com", "test_service", rand.Float64()*10)
		if c.TrackEvent(e) {
			c.DoOnTracker(e, func(et *eventTracker) {
				t.Fatal(et.df.Data())
			})
		}
	}
	e := newTestEvent("machine.test.com", "test_service", 1000000000.0)
	if !c.TrackEvent(e) {
		c.DoOnTracker(e, func(et *eventTracker) {
			t.Fatal(et.df.Data())
		})
	}

}
