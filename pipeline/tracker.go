package pipeline

import (
	"sync/atomic"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/event"
)

var (
	EMPTY_SUB_SERVICE = ""
)

// Provides stat tracking for events
type Tracker struct {
	inChan      chan *event.Event
	queryChan   chan QueryFunc
	total       *counter
	hosts       map[string]*counter
	services    map[string]*counter
	subServices map[string]*counter
}

// holds information about the current state of an event tracker
type TrackerReport struct {
	Total        uint64            `json:"total_events"`
	ByHost       map[string]uint64 `json:"by_host"`
	ByService    map[string]uint64 `json:"by_service"`
	BySubService map[string]uint64 `json:"by_sub_service"`
}

func NewReport() *TrackerReport {
	return &TrackerReport{
		ByHost:       make(map[string]uint64),
		ByService:    make(map[string]uint64),
		BySubService: make(map[string]uint64),
	}
}

// create and return a new *Tracker
func NewTracker() *Tracker {
	t := &Tracker{
		inChan:      make(chan *event.Event),
		queryChan:   make(chan QueryFunc),
		total:       &counter{},
		hosts:       make(map[string]*counter),
		services:    make(map[string]*counter),
		subServices: make(map[string]*counter),
	}

	return t
}

// return a report of the current state of the tracker
func (t *Tracker) GetStats() *TrackerReport {
	logrus.Info("Fetching stats")
	r := NewReport()
	t.Query(func(t *Tracker) {
		r.Total = t.total.get()
		for k, v := range t.hosts {
			r.ByHost[k] = v.get()
		}
		for k, v := range t.services {
			r.ByService[k] = v.get()
		}
		for k, v := range t.subServices {
			r.BySubService[k] = v.get()
		}
	})

	return r
}

// An function that is given access to the tracker without locks
type QueryFunc func(t *Tracker)

// Start the tracker. This should be done in it's own goroutine
func (t *Tracker) Start() {
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

// Perform a QueryFunc on the tracker syncronously
func (t *Tracker) Query(f QueryFunc) {
	done := make(chan struct{})
	t.queryChan <- func(t *Tracker) {
		f(t)
		done <- struct{}{}
	}
	<-done
}

func (t *Tracker) query(f QueryFunc) {
	f(t)
}

func (t *Tracker) trackEvent(e *event.Event) {
	t.total.inc()

	// increment host counter
	host, ok := t.hosts[e.Host]
	if !ok {
		host = &counter{}
		t.hosts[e.Host] = host
	}
	host.inc()

	// increment service counter
	service, ok := t.services[e.Service]
	if !ok {
		service = &counter{}
		t.services[e.Service] = service
	}
	service.inc()

	// if the event has a sub_service, increment the sub_service counter
	if e.SubService != EMPTY_SUB_SERVICE {
		subService, ok := t.subServices[e.SubService]
		if !ok {
			subService = &counter{}
			t.subServices[e.SubService] = subService
		}

		subService.inc()
	}
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
