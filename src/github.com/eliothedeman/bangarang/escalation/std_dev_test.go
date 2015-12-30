package escalation

import (
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
			c.DoOnTracker(e, func(et *eventTracker) {
				t.Fatal(et.df.Data())
			})
		}
	}
}

func TestStdDevQuarter(t *testing.T) {
	c := &Condition{
		Greater:    test_f(1),
		StdDev:     true,
		WindowSize: 100,
	}

	c.init(DEFAULT_GROUP_BY)

	// insert less than 1/4 of the window size
	for i := 0; i < 20; i++ {
		e := newTestEvent("machine.test.com", "test_service", rand.Float64()*10)
		if c.TrackEvent(e) {
			t.Fatal(c.getTracker(e).df.Data())
		}
	}

	// now insert something that is wayyyy out of the std_dev
	e := newTestEvent("machine.test.com", "test_service", 10000000.0)
	if c.TrackEvent(e) {
		c.DoOnTracker(e, func(et *eventTracker) {
			t.Fatal(et.df.Data())
		})
	}
}

func TestStdDev(t *testing.T) {
	c := &Condition{
		Greater:    test_f(1),
		StdDev:     true,
		WindowSize: 100,
	}

	c.init(DEFAULT_GROUP_BY)

	for i := 0; i < 25; i++ {
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
