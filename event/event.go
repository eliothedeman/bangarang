package event

import (
	"fmt"
	"sync"
	"time"
)

const (
	OK       = 0
	WARNING  = 1
	CRITICAL = 2
)

var (
	DEFAULT_TTL = time.Minute * 10
)

type Event struct {
	Host       string            `json:"host"`
	Service    string            `json:"host"`
	SubService string            `json:"sub_type"`
	Metric     float64           `json:"metric"`
	Time       time.Time         `json:"time"`
	Occurences int               `json:"occurences"`
	Tags       map[string]string `json:"tags"`
	TTL        time.Duration     `json:"-"`
	Status     int               `json:"status"`
}

func (e *Event) IndexName() string {
	return fmt.Sprintf("%s:%s:%s", e.Host, e.Service, e.SubService)
}

type Index struct {
	events map[string]*Event
	sync.RWMutex
}

func (i *Index) Put(e *Event) {
	i.Lock()
	i.events[e.IndexName()] = e
	i.Unlock()
}

func (i *Index) Get(n string) *Event {
	i.RLock()
	e, ok := i.events[n]
	i.RUnlock()
	if !ok {
		return nil
	}
	return e
}

func (i *Index) GetByAge(age time.Duration) []*Event {
	now := time.Now()
	i.RLock()
	events := make([]*Event, 0, len(i.events)/10)
	for _, e := range i.events {
		if now.Sub(e.Time) > age {
			events = append(events, e)
		}
	}
	i.RUnlock()
	return events
}
