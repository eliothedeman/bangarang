package alarm

import "testing"

func TestDerivativeFalsePositive(t *testing.T) {
	c := &Condition{
		Greater:    test_f(10),
		Derivative: true,
		WindowSize: 1,
	}

	c.init(DEFAULT_GROUP_BY)

	// make sure derivatives below 10 don't satisfy the condition
	for i := 0; i < 300; i++ {
		e := newTestEvent("machine.test.com", "test_service", float64(i*5))
		if c.TrackEvent(e) {
			c.DoOnTracker(e, func(et *eventTracker) {
				t.Fatal(et.df.Data())
			})
		}
	}
}

func TestDerivative(t *testing.T) {
	c := &Condition{
		Greater:    test_f(100),
		Derivative: true,
		Occurences: 1,
		WindowSize: 1,
	}

	c.init(DEFAULT_GROUP_BY)
	e := newTestEvent("machine.test.com", "test_service", float64(10))
	if c.TrackEvent(e) {
		t.Fatal("Derivative checks must see more than 1 value before being met")
	}

	e = newTestEvent("machine.test.com", "test_service", float64(111))
	if !c.TrackEvent(e) {
		t.Fatal(c.getTracker(e).df.Data())
	}

	// make sure it will resolve
	e = newTestEvent("machine.test.com", "test_service", float64(112))
	if c.TrackEvent(e) {
		t.Fatal()
	}
}

func TestNegativeDerivative(t *testing.T) {
	c := &Condition{
		Less:       test_f(-10),
		Derivative: true,
		Occurences: 1,
		WindowSize: 1,
	}

	c.init(DEFAULT_GROUP_BY)
	e := newTestEvent("machine.test.com", "test_service", float64(10))
	if c.TrackEvent(e) {
		t.Fatal("Derivative checks must see more than 1 value before being met")
	}

	e = newTestEvent("machine.test.com", "test_service", float64(-4))
	if !c.TrackEvent(e) {
		t.Fatal()
	}

	// make sure it will resolve
	e = newTestEvent("machine.test.com", "test_service", float64(112))
	if c.TrackEvent(e) {
		t.Fatal()
	}

}
