package event

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"unsafe"
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
	if len(buff) != int(binary.BigEndian.Uint64(buff[:8])) {
		return fmt.Errorf("Unmarshal Binary: Malformed binary blob. Expected length of %d got %d", buff[0], len(buff))
	}

	// load the metric's bits as a uint64
	i := binary.BigEndian.Uint64(buff[8:16])
	e.Metric = math.Float64frombits(i)

	// host
	offset := 16
	l := int(buff[offset])
	offset += 1
	e.Host = string(buff[offset : offset+l])
	offset += l

	// service
	l = int(buff[offset])
	offset += 1
	e.Service = string(buff[offset : offset+l])
	offset += l

	// sub_service
	l = int(buff[offset])
	offset += 1
	e.SubService = string(buff[offset : offset+l])
	offset += l

	// tags
	e.Tags = map[string]string{}
	for offset < len(buff) {

		// key
		l = int(buff[offset])
		offset += 1
		key := string(buff[offset : offset+l])
		offset += l

		l = int(buff[offset])
		offset += 1
		value := string(buff[offset : offset+l])
		offset += l

		e.Tags[key] = value
	}

	return nil
}

// MarshalBinary creates the binary representation of an event
// Size header 8 bytes
// Metric 8 bytes
// Host Size 1 byte
// Host (max 256 bytes)
// Service Size 1 byte
// Service (max 256 bytes)
// SubService Size 1 byte
// SubService (max 256 bytes)
//
// Tags: key size 1 byte
// Tags: key (max 256 bytes)
// Tags: value size 1 byte
// Tags: value (max 256 bytes)
// repeat
func (e *Event) MarshalBinary() ([]byte, error) {
	size := e.binSize()
	buff := make([]byte, size)
	offset := 0
	tmp := 0

	// size
	binary.BigEndian.PutUint64(buff[:8], uint64(size))
	offset += 8

	// metric
	binary.BigEndian.PutUint64(buff[offset:offset+8], math.Float64bits(e.Metric))
	offset += 8

	// host
	tmp = sizeOfString(e.Host)
	buff[offset] = uint8(tmp)
	offset += 1
	copy(buff[offset:offset+tmp], unsafeBytes(e.Host))
	offset += tmp

	// service
	tmp = sizeOfString(e.Service)
	buff[offset] = uint8(tmp)
	offset += 1
	copy(buff[offset:offset+tmp], unsafeBytes(e.Service))
	offset += tmp

	// service
	tmp = sizeOfString(e.SubService)
	buff[offset] = uint8(tmp)
	offset += 1
	copy(buff[offset:offset+tmp], unsafeBytes(e.SubService))
	offset += tmp

	// tags
	for k, v := range e.Tags {
		tmp = sizeOfString(k)
		buff[offset] = uint8(tmp)
		offset += 1
		copy(buff[offset:offset+tmp], unsafeBytes(k))
		offset += tmp

		tmp = sizeOfString(v)
		buff[offset] = uint8(tmp)
		offset += 1
		copy(buff[offset:offset+tmp], unsafeBytes(v))
		offset += tmp
	}

	return buff, nil
}

// binSize returns the size of an event once encoded as binary
func (e *Event) binSize() int {

	// start with the size of the "size" header + size of metric
	size := 16

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

func unsafeBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Len:  len(s),
		Cap:  len(s),
		Data: (*(*reflect.StringHeader)(unsafe.Pointer(&s))).Data,
	}))
}

// return the size of all the string in the map
func sizeOfMap(m map[string]string) int {
	// key/val headers
	size := len(m) * 2
	for k, v := range m {

		// size of key
		size += sizeOfString(k)

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
