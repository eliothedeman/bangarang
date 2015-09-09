package event

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
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
}

func (e *Event) UnmarshalBinary(buff []byte) error {
	if len(buff) != int(buff[0]) {
		return fmt.Errorf("Unmarshal Binary: Malformed binary blob. Expected length of %d got %d", buff[0], len(buff))
	}

	// load the metric's bits as a uint64
	i, _ := binary.Uvarint(buff[8:16])
	e.Metric = math.Float64frombits(i)

	return nil
}

// MarshalBinary creates the binary representation of an event
// Size header 8 bytes
// Metric 8 bytes
// Host Size 1 byte
// Host (max 256 bytes)
// Service Size 1 byte
// Service (max 256 bytes)
func (e *Event) MarshalBinary() ([]byte, error) {
	size := e.binSize()
	buff := make([]byte, size)
	offset := 0

	// size
	binary.PutUvarint(buff[:8], uint64(size))
	offset += 8

	// metric
	log.Println(buff[offset : offset+8])
	binary.PutUvarint(buff[offset:offset+8], math.Float64bits(e.Metric))
	offset += 8

	// host
	buff[offset] = uint8(sizeOfString(e.Host))
	offset += 1
	copy(buff[offset:offset+sizeOfString(e.Host)], []byte(e.Host))
	offset += sizeOfString(e.Host)

	// service
	buff[offset] = uint8(sizeOfString(e.Service))
	offset += 1
	copy(buff[offset:offset+sizeOfString(e.Service)], []byte(e.Service))
	offset += sizeOfString(e.Service)

	// service
	buff[offset] = uint8(sizeOfString(e.SubService))
	offset += 1
	copy(buff[offset:offset+sizeOfString(e.SubService)], []byte(e.SubService))
	offset += sizeOfString(e.SubService)

	// tags
	for k, v := range e.Tags {
		buff[offset] = uint8(sizeOfString(k))
		offset += 1
		copy(buff[offset:offset+sizeOfString(k)], []byte(k))
		offset += sizeOfString(k)

		buff[offset] = uint8(sizeOfString(v))
		offset += 1
		copy(buff[offset:offset+sizeOfString(v)], []byte(v))
		offset += sizeOfString(v)
	}

	return buff, nil
}

// binSize returns the size of an event once encoded as binary
func (e *Event) binSize() int {

	// start with the size of the "size" header
	size := 8

	// add the size of the metric
	size += 8

	// all non-fixed sizes also have an 1 byte size field

	// get the size of all the strings
	size += sizeOfString(e.Host)
	size += 1

	size += sizeOfString(e.Service)
	size += 1

	size += sizeOfString(e.SubService)
	size += 1

	// get the size of the tags
	size += sizeOfMap(e.Tags)

	return size
}

func sizeOfString(s string) int {
	size := len(s)
	if size > 256 {
		return 256
	}

	return size
}

// return the size of all the string in the map
func sizeOfMap(m map[string]string) int {
	size := 0
	for k, v := range m {
		// header
		size += 1

		// size of key
		size += sizeOfString(k)

		// val header
		size += 1

		// size of value
		size += sizeOfString(v)
	}
	return size
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
