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
	Service    string            `json:"service"`
	SubService string            `json:"sub_type"`
	Metric     float64           `json:"metric"`
	Time       time.Time         `json:"time"`
	Occurences int               `json:"occurences"`
	Tags       map[string]string `json:"tags"`
	TTL        time.Duration     `json:"-"`
	Status     int               `json:"status"`
	LastEvent  *Event            `json:"-"`
}

func (e *Event) IndexName() string {
	return fmt.Sprintf("%s:%s:%s", e.Host, e.Service, e.SubService)
}

func (e *Event) StatusChanged() bool {
	if e.LastEvent == nil {
		return e.Status != OK
	}

	return !(e.LastEvent.Status == e.Status)
}

func status(code int) string {
	switch code {
	case WARNING:
		return "warning"
	case CRITICAL:
		return "critical"
	default:
		return "ok"
	}
}

func (e *Event) FormatDescription() string {
	return fmt.Sprintf("%s! %s.%s on %s is %.2f", status(e.Status), e.Service, e.SubService, e.Host, e.Metric)
}

type Index struct {
	events     map[string]*Event
	keepAlives map[string]time.Time
	sync.RWMutex
}

func NewIndex() *Index {
	return &Index{
		events:     make(map[string]*Event),
		keepAlives: make(map[string]time.Time),
	}
}

// get all of the hosts that have missed their keepalives
func (i *Index) GetExpired(age time.Duration) []string {
	hosts := make([]string, 0, 10)
	n := time.Now()
	i.RLock()
	for host, t := range i.keepAlives {
		if n.Sub(t) > age {
			hosts = append(hosts, host)
		}
	}
	i.RUnlock()
	return hosts
}

func (i *Index) Put(e *Event) {
	name := e.IndexName()
	e.LastEvent = i.Get(name)
	i.Lock()
	i.events[name] = e
	i.keepAlives[e.Host] = time.Now()
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
