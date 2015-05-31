package alarm

import (
	"math"
	"sync"

	"github.com/eliothedeman/bangarang/event"
	"github.com/eliothedeman/smoothie"
)

var (
	DEFAULT_WINDOW_SIZE = 100
	STATUS_SIZE         = 10
)

type Condition struct {
	Greater       *float64 `json:"greater"`
	Less          *float64 `json:"less"`
	Exactly       *float64 `json:"exactly"`
	StdDev        *StdDev  `json:"std_dev"`
	Escalation    string   `json:"escalation"`
	Occurences    int      `json:"occurences"`
	WindowSize    int      `json:"window_size"`
	checks        []satisfier
	eventTrackers map[string]*eventTracker
	sync.Mutex
	ready bool
}

type eventTracker struct {
	df         *smoothie.DataFrame
	states     *smoothie.DataFrame
	occurences int
}

type satisfier func(e *event.Event) bool

type StdDev struct {
	Sigma      float64 `json:"sigma"`
	WindowSize *int    `json:"window_size"`
}

func (c *Condition) DoOnTracker(e *event.Event, dot func(*eventTracker)) {
	c.Lock()
	et, ok := c.eventTrackers[e.IndexName()]
	if !ok {
		df := smoothie.NewDataFrame(c.WindowSize)
		states := smoothie.NewDataFrameFromSlice(make([]float64, STATUS_SIZE))
		et = &eventTracker{
			df:     df,
			states: states,
		}
		c.eventTrackers[e.IndexName()] = et
	}
	dot(et)
	c.Unlock()
}

// start tracking an event, and returns if the event has hit it's occurence settings
func (c *Condition) TrackEvent(e *event.Event) bool {
	c.DoOnTracker(e, func(t *eventTracker) {
		t.df.Push(e.Metric)
	})

	return c.OccurencesHit(e)

}

func (c *Condition) StateChanged(e *event.Event) bool {
	changed := false
	c.DoOnTracker(e, func(t *eventTracker) {
		changed = t.states.Index(0) == t.states.Index(1)
	})
	return changed
}

// check to see if an event has it the occurences level
func (c *Condition) OccurencesHit(e *event.Event) bool {
	occ := 0

	if c.Satisfies(e) {
		c.DoOnTracker(e, func(t *eventTracker) {
			t.occurences += 1
			occ = t.occurences
			t.states.Push(1)
		})
	} else {
		c.DoOnTracker(e, func(t *eventTracker) {
			t.occurences = 0
			t.states.Push(0)
		})
	}

	return occ >= c.Occurences
}

// check if an event satisfies a condition
func (c *Condition) Satisfies(e *event.Event) bool {
	for _, check := range c.checks {
		if check(e) {
			return true
		}
	}

	return false
}

// create a list of checks that the condition will use to test events
func (c *Condition) compileChecks() []satisfier {
	s := []satisfier{}

	if c.Greater != nil {
		s = append(s, func(e *event.Event) bool {
			return e.Metric > *c.Greater
		})
	}
	if c.Less != nil {
		s = append(s, func(e *event.Event) bool {
			return e.Metric < *c.Less
		})
	}
	if c.Exactly != nil {
		s = append(s, func(e *event.Event) bool {
			return e.Metric == *c.Less
		})
	}
	if c.StdDev != nil {
		s = append(s, func(e *event.Event) bool {
			met := false
			c.DoOnTracker(e, func(t *eventTracker) {
				avg := t.df.Avg()
				if math.Abs(e.Metric-avg) > t.df.StdDev()*c.StdDev.Sigma {
					met = true
				}
			})
			return met
		})
	}
	return s
}

func (c *Condition) init() {
	c.checks = c.compileChecks()

	// fixes issue where occurences are hit, even when the event doesn't satisify the condition
	if c.Occurences < 1 {
		c.Occurences = 1
	}

	if c.eventTrackers == nil {
		c.eventTrackers = make(map[string]*eventTracker)
	}

	if c.WindowSize == 0 {
		c.WindowSize = DEFAULT_WINDOW_SIZE
	}
	c.ready = true
}
