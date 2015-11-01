package pipeline

import (
	"log"
	"sync/atomic"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/event"
)

var (
	EMPTY_SUB_SERVICE = ""
)

const (
	INTERNAL_TAG_NAME = ""
)

// Provides stat tracking for events
type Tracker struct {
	inChan            chan *event.Event
	started           atomic.Value
	queryChan         chan QueryFunc
	total             *counter
	totalIncidents    counter
	incidentResolvers map[string]chan *event.Incident
	tagCounters       map[string]map[string]*counter
	tagTimers         map[string]map[string]time.Time
}

func (t *Tracker) Started() bool {
	if t.started.Load() == nil {
		return false
	} else {
		return t.started.Load().(bool)
	}
}

// create and return a new *Tracker
func NewTracker() *Tracker {
	t := &Tracker{
		inChan:            make(chan *event.Event),
		queryChan:         make(chan QueryFunc),
		total:             &counter{},
		incidentResolvers: make(map[string]chan *event.Incident),
		tagCounters:       make(map[string]map[string]*counter),
		tagTimers:         make(map[string]map[string]time.Time),
	}

	return t
}

// holds information about the current state of an event tracker
type TrackerReport struct {
	Total         uint64                       `json:"total_events"`
	CountByTag    map[string]map[string]uint64 `json:"count_by_tag"`
	LastSeenByTag map[string]map[string]int64  `json:"last_seen_by_tag"`
}

func NewReport() *TrackerReport {
	return &TrackerReport{
		CountByTag:    make(map[string]map[string]uint64),
		LastSeenByTag: make(map[string]map[string]int64),
	}
}

// TrackIncident will allow the tracker to keep state about an incident
func (t *Tracker) TrackIncident(i *event.Incident) {
	if i.GetResolve() != nil {
		t.Query(func(r *Tracker) {

			// Don't keep track of "OK" incident resolvers, as ok's can't be resolved
			if i.Status != event.OK {
				r.incidentResolvers[string(i.IndexName())] = i.GetResolve()
			}
			r.totalIncidents.inc()
		})
	}
}

func (t *Tracker) GetIncidentResolver(i *event.Incident) chan *event.Incident {
	var res chan *event.Incident
	t.Query(func(r *Tracker) {
		res, _ = r.incidentResolvers[string(i.IndexName())]
	})
	return res
}

// return a report of the current state of the tracker
func (t *Tracker) GetStats() *TrackerReport {
	logrus.Info("Fetching stats")
	r := NewReport()
	t.Query(func(t *Tracker) {
		r.Total = t.total.get()

		for tag, data := range t.tagCounters {
			tmp := make(map[string]uint64, len(data))
			for k, v := range data {
				tmp[k] = v.get()
			}
			r.CountByTag[tag] = tmp
		}

		for tag, data := range t.tagTimers {
			tmp := make(map[string]int64, len(data))
			for k, v := range data {
				tmp[k] = v.Unix()
			}
			r.LastSeenByTag[tag] = tmp
		}

	})

	return r
}

// ListTags
func (t *Tracker) ListTags() []string {
	var tags []string
	t.Query(func(t *Tracker) {
		tags := make([]string, 0, len(t.tagTimers))
		for k, _ := range t.tagTimers {
			tags = append(tags, k)
		}
	})
	return tags
}

func (t *Tracker) GetTag(tag string) []string {
	var tags []string
	t.Query(func(t *Tracker) {
		tmp, ok := t.tagCounters[tag]

		// stop short if we have never seen this tag before
		if !ok {
			tags = []string{}
			return
		}

		// get them tags
		tags = make([]string, len(tmp))
		x := 0
		for k, _ := range tmp {
			tags[x] = k
			x += 1
		}
	})

	return tags
}

func (t *Tracker) RemoveTag(tag, key string) {
	t.QueryAsync(func(t *Tracker) {

		// remove counters
		if tmp, ok := t.tagCounters[tag]; ok {
			delete(tmp, key)
		}

		// remove timers
		if tmp, ok := t.tagTimers[tag]; ok {
			delete(tmp, key)

		}
	})
}

func (t *Tracker) TagTimes(tag string) map[string]time.Time {
	m := make(map[string]time.Time)
	t.Query(func(t *Tracker) {
		if timers, ok := t.tagTimers[tag]; ok {
			for k, v := range timers {
				m[k] = v
			}
		}
	})

	return m
}

// An function that is given access to the tracker without locks
type QueryFunc func(t *Tracker)

// Start the tracker. This should be done in it's own goroutine
func (t *Tracker) Start() {
	t.started.Store(true)
	logrus.Info("Starting event tracker")
	var e *event.Event
	var f QueryFunc
	for {
		select {
		case e = <-t.inChan:
			t.trackEvent(e)
		case f = <-t.queryChan:
			t.query(f)
		}
	}
}

// Process the stats for a given event
func (t *Tracker) TrackEvent(e *event.Event) {
	t.inChan <- e
}

// Query exicutes the given QueryFunc ayncronously
func (t *Tracker) Query(f QueryFunc) {
	done := make(chan struct{})
	t.queryChan <- func(t *Tracker) {
		f(t)
		done <- struct{}{}
	}
	<-done
}

// QueryAsync exicutes the given QueryFunc asyncronously
func (t *Tracker) QueryAsync(f QueryFunc) {
	t.queryChan <- f
}

func (t *Tracker) query(f QueryFunc) {
	f(t)
}

func (t *Tracker) updateTimes(e *event.Event) {
	now := time.Now()
	e.Tags.ForEach(func(k, v string) {
		tmp, ok := t.tagTimers[k]
		if !ok {
			tmp = make(map[string]time.Time)
			t.tagTimers[k] = tmp
		}

		tmp[v] = now
	})
}

func (t *Tracker) updateCounts(e *event.Event) {
	log.Println(e.Tags)

	e.Tags.ForEach(func(k, v string) {
		println(k, v)
		tmp, ok := t.tagCounters[k]
		if !ok {
			tmp = make(map[string]*counter)
			t.tagCounters[k] = tmp
		}

		c, ok := tmp[v]
		if !ok {
			c = &counter{}
			tmp[v] = c
		}
		c.inc()
	})
}

func (t *Tracker) trackEvent(e *event.Event) {
	// signal that this event has been tracked
	defer e.WaitDec()

	// don't track internal events
	if len(e.Get(INTERNAL_TAG_NAME)) != 0 {
		return
	}
	t.total.inc()

	t.updateCounts(e)
	t.updateTimes(e)

}

type counter struct {
	c uint64
}

func (c *counter) inc() {
	atomic.AddUint64(&c.c, 1)
}
func (c *counter) set(val uint64) {
	atomic.StoreUint64(&c.c, val)
}

func (c *counter) get() uint64 {
	return atomic.LoadUint64(&c.c)
}
