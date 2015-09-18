package event

import (
	"encoding/binary"
	"strings"
	"sync"
)

const (
	OK = iota
	WARNING
	CRITICAL
)

//go:generate ffjson $GOFILE
//go:generate msgp $GOFILE

// Event represents a metric as it passes through the pipeline.
// it holds meta data about the metric, as well as methods to trace the event as it is processed
type Event struct {
	Host       string            `json:"host" msg:"host"`
	Service    string            `json:"service" msg:"service"`
	SubService string            `json:"sub_service" msg:"sub_service"`
	Metric     float64           `json:"metric" msg:"metric"`
	Tags       map[string]string `json:"tags" msg:"tags"`
	indexName  string
	wait       sync.WaitGroup
	sync.Mutex
}

func (e *Event) MarshalBinary() ([]byte, error) {
	buff := make([]byte, e.Msgsize()+8)
	tmp, err := e.MarshalJSON()
	if err != nil {
		return nil, err
	}
	copy(buff[8:], tmp)

	binary.PutUvarint(buff[:8], uint64(len(tmp)))
	return buff[:8+len(tmp)], nil
}

func (e *Event) Wait() {
	e.wait.Wait()
}

// WaitDec decrements the event's waitgroup counter
func (e *Event) WaitDec() {
	e.Lock()
	e.wait.Done()
	e.Unlock()
}

// WaitAdd increments ot the event's waitgroup counter
func (e *Event) WaitInc() {
	e.Lock()
	e.wait.Add(1)
	e.Unlock()
}

// Passer provides a method for passing an event down a step in the pipeline
type Passer interface {
	Pass(e *Event)
}

func NewEvent() *Event {
	e := &Event{}
	return e
}

// Get any value on an event as a string
func (e *Event) Get(key string) string {

	// attempt to find the string values of the event
	switch strings.ToLower(key) {
	case "host":
		return e.Host
	case "service":
		return e.Service
	case "sub_service":
		return e.SubService
	}

	// if we make it to this point, assume we are looking for a tag
	if val, ok := e.Tags[key]; ok {
		return val
	}

	return ""
}

func (e *Event) IndexName() string {
	return e.Host + e.Service + e.SubService
}

func Status(code int) string {
	switch code {
	case WARNING:
		return "warning"
	case CRITICAL:
		return "critical"
	default:
		return "ok"
	}
}
