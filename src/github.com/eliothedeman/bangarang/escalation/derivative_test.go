package escalation

import (
	"testing"

	"github.com/eliothedeman/bangarang/event"
)

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

func TestDerivativeWindow(t *testing.T) {
	var tests = []struct {
		desc string
		c    *Condition
		e    []*event.Event
		want []bool
	}{
		{
			desc: "two events, large window size",
			c: &Condition{
				Greater:    test_f(100),
				Derivative: true,
				Occurences: 1,
				WindowSize: 100,
			},
			e: []*event.Event{
				&event.Event{
					Tags: &event.TagSet{
						{"host", "machine.test.com"},
						{"service", "test.service"},
					},
					Metric: 0,
				},
				&event.Event{
					Tags: &event.TagSet{
						{"host", "machine.test.com"},
						{"service", "test.service"},
					},
					Metric: 1000,
				},
			},
			want: []bool{false, false},
		},
		{
			desc: "hit window size",
			c: &Condition{
				Greater:    test_f(100),
				Derivative: true,
				Occurences: 1,
				WindowSize: 2,
			},
			e: []*event.Event{
				&event.Event{
					Tags: &event.TagSet{
						{"host", "machine.test.com"},
						{"service", "test.service"},
					},
					Metric: 0,
				},
				&event.Event{
					Tags: &event.TagSet{
						{"host", "machine.test.com"},
						{"service", "test.service"},
					},
					Metric: 1000,
				},
			},
			want: []bool{false, true},
		},
	}

	for i, tt := range tests {
		tt.c.init(DEFAULT_GROUP_BY)
		for x, e := range tt.e {
			got := tt.c.TrackEvent(e)
			if tt.want[x] != got {
				t.Fatalf("%d test: %s wanted %t got %t", i, tt.desc, tt.want[x], got)
			}
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
