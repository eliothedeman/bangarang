package alarm

import (
	"sync"
	"sync/atomic"

	"github.com/eliothedeman/bangarang/event"
)

type Condition struct {
	Greater      *float64                 `json:"greater"`
	Less         *float64                 `json:"less"`
	Exactly      *float64                 `json:"exactly"`
	Occurences   int                      `json:"occurences"`
	tracker      map[string]*EventTracker `-`
	trackerMutex sync.RWMutex
}

// start tracking an event, and returns if the event has hit it's occurence settings
func (c *Condition) TrackEvent(e *event.Event) bool {
	c.initTracker()
	t := c.getEventTracker(e)
	t.Inc()
	return t.Occurences() >= c.Occurences
}

func (c *Condition) CleanEvent(e *event.Event) {
	c.initTracker()
	c.trackerMutex.Lock()
	delete(c.tracker, e.FormatDescription())
	c.trackerMutex.Unlock()

}

func (c *Condition) getEventTracker(e *event.Event) *EventTracker {
	c.trackerMutex.RLock()
	t, ok := c.tracker[e.IndexName()]
	c.trackerMutex.RUnlock()
	if !ok {
		t = NewEventTracker(e)
		c.trackerMutex.Lock()
		c.tracker[e.IndexName()] = t
		c.trackerMutex.Unlock()
	}

	return t
}

func (c *Condition) initTracker() {
	c.trackerMutex.RLock()
	if c.tracker == nil {
		c.tracker = make(map[string]*EventTracker)
	}
	c.trackerMutex.RUnlock()
}

type EventTracker struct {
	event     *event.Event
	occurence int64
}

func NewEventTracker(e *event.Event) *EventTracker {
	return &EventTracker{
		event: e,
	}
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

	return false
}
