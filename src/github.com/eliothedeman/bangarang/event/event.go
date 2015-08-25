package event

import (
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

type Event struct {
	Host       string            `json:"host" msg:"host"`
	Service    string            `json:"service" msg:"service"`
	SubService string            `json:"sub_service" msg:"sub_service"`
	Metric     float64           `json:"metric" msg:"metric"`
	Occurences int               `json:"occurences" msg:"occurences"`
	Tags       map[string]string `json:"tags" msg:"tags"`
	Id         uint64
	indexName  string
	wait       sync.WaitGroup
}

func (e *Event) Wait() {
	e.wait.Wait()
}

// WaitDec decrements the event's waitgroup counter
func (e *Event) WaitDec() {
	e.wait.Done()
}

// WaitAdd increments ot the event's waitgroup counter
func (e *Event) WaitInc() {
	e.wait.Add(1)
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
	if len(e.indexName) == 0 {
		e.indexName = e.Host + e.Service + e.SubService
	}
	return e.indexName
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
