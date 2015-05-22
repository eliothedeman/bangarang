package alarm

import (
	"math"
	"sync"
	"sync/atomic"

	"github.com/eliothedeman/bangarang/event"
	"github.com/eliothedeman/smoothie"
)

//go:generate ffjson $GOFILE

var (
	DEFAULT_WINDOW_SIZE = 100
)

type Condition struct {
	Greater      *float64                 `json:"greater"`
	Less         *float64                 `json:"less"`
	Exactly      *float64                 `json:"exactly"`
	StdDev       *StdDev                  `json:"std_dev"`
	Occurences   int                      `json:"occurences"`
	tracker      map[string]*EventTracker `-`
	trackerMutex sync.RWMutex
}

type StdDev struct {
	Sigma      float64 `json:"sigma"`
	WindowSize *int    `json:"window_size"`
}

// start tracking an event, and returns if the event has hit it's occurence settings
func (c *Condition) TrackEvent(e *event.Event) bool {
	c.initTracker()
	t := c.getEventTracker(e)

	if c.trackingStats() {
		t.df.Push(e.Metric)
	}

	if c.Satisfies(e) {
		t.Inc()
		return t.Occurences() >= c.Occurences
	}

	return false

}

func (c *Condition) CleanEvent(e *event.Event) {
	c.initTracker()
	c.trackerMutex.Lock()
	delete(c.tracker, e.FormatDescription())
	c.trackerMutex.Unlock()

}

func (c *Condition) trackingStats() bool {
	return c.StdDev != nil
}

func (c *Condition) getEventTracker(e *event.Event) *EventTracker {
	c.trackerMutex.RLock()
	t, ok := c.tracker[e.IndexName()]
	c.trackerMutex.RUnlock()
	if !ok {
		t = NewEventTracker()
		if c.trackingStats() {
			if c.StdDev.WindowSize == nil {
				c.StdDev.WindowSize = &DEFAULT_WINDOW_SIZE
			}
			t.initDataFrame(*c.StdDev.WindowSize)
		}
		c.trackerMutex.Lock()
		c.tracker[e.IndexName()] = t
		c.trackerMutex.Unlock()
	}

	return t
}

func (c *Condition) initTracker() {
	if c.tracker == nil {
		c.trackerMutex.Lock()
		c.tracker = make(map[string]*EventTracker)
		c.trackerMutex.Unlock()
	}
}

type EventTracker struct {
	df        *smoothie.DataFrame
	occurence int64
}

func (e *EventTracker) initDataFrame(windowSize int) {
	e.df = smoothie.EmptyDataFrame(windowSize)
}

func NewEventTracker() *EventTracker {
	return &EventTracker{}
}

func (e *EventTracker) Inc() {
	atomic.AddInt64(&e.occurence, 1)
}

func (e *EventTracker) Reset() {
	atomic.StoreInt64(&e.occurence, 0)
}

func (e *EventTracker) Occurences() int {
	return int(atomic.LoadInt64(&e.occurence))
}

// check if an event satisfies a condition
func (c *Condition) Satisfies(e *event.Event) bool {
	if c.Greater != nil {
		if e.Metric > *c.Greater {
			return true
		}
	}

	if c.Less != nil {
		if e.Metric < *c.Less {
			return true
		}
	}

	if c.Exactly != nil {
		if e.Metric == *c.Exactly {
			return true
		}
	}

	if c.trackingStats() {

		// check if the current metric is outside of n * sigma
		if c.StdDev != nil {
			c.trackerMutex.RLock()
			t := c.tracker[e.IndexName()]
			avg := t.df.Avg()
			if math.Abs(e.Metric-avg) > t.df.StdDev()*c.StdDev.Sigma {
				return true
			}
			c.trackerMutex.RUnlock()
		}
	}

	return false
}
