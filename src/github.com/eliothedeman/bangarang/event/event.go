package event

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"sync"
	"time"
	"unsafe"
)

const (
	OK = iota
	WARNING
	CRITICAL
)

// Event represents a metric as it passes through the pipeline.
// It holds meta data about the metric, as well as methods to trace the event as it is processed
type Event struct {
	Metric    float64   `json:"metric" msg:"metric"`
	Tags      *TagSet   `json:"tags" msg:"tags"`
	Time      time.Time `json:"time" msg: "time"`
	indexName string
	wait      sync.WaitGroup
	mut       sync.Mutex
}

func (e *Event) UnmarshalBinary(buff []byte) error {
	if len(buff) != int(binary.BigEndian.Uint64(buff[:8])) {
		return fmt.Errorf("Unmarshal Binary: Malformed binary blob. Expected length of %d got %d", buff[0], len(buff))
	}

	// load the metric's bits as a uint64
	i := binary.BigEndian.Uint64(buff[8:16])
	e.Metric = math.Float64frombits(i)

	// load the time
	// seconds
	i = binary.BigEndian.Uint64(buff[16:24])

	// nanoseconds
	x := binary.BigEndian.Uint64(buff[24:32])
	e.Time = time.Unix(int64(i), int64(x))

	l := 0
	offset := 32

	// tags
	e.Tags = NewTagset(int(buff[offset]))
	offset += 1
	i = 0
	var key, val string
	for offset < len(buff) {

		// key
		l = int(buff[offset])
		offset += 1
		key = string(buff[offset : offset+l])
		offset += l

		l = int(buff[offset])
		offset += 1
		val = string(buff[offset : offset+l])
		offset += l

		e.Tags.Set(key, val)

		i += 1

	}

	return nil
}

// MarshalBinary creates the binary representation of an event
// Size header 8 bytes
// Metric 8 bytes
// Time seconds 8 bytes
// Time nanoseconds  8 bytes
//
// TagSetSize 1 byte (max 256 key->val pairs)
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

	// time
	// seconds
	binary.BigEndian.PutUint64(buff[offset:offset+8], uint64(e.Time.Unix()))
	offset += 8

	// nano seconds
	binary.BigEndian.PutUint64(buff[offset:offset+8], uint64(e.Time.Nanosecond()))
	offset += 8

	buff[offset] = uint8(len(*e.Tags))
	offset += 1

	// tags
	e.Tags.ForEach(func(k, v string) {
		tmp = sizeOfString(k)
		buff[offset] = uint8(tmp)
		offset += 1
		copy(buff[offset:offset+tmp], k)
		offset += tmp

		tmp = sizeOfString(v)
		buff[offset] = uint8(tmp)
		offset += 1
		copy(buff[offset:offset+tmp], v)
		offset += tmp

	})

	return buff, nil
}

// binSize returns the size of an event once encoded as binary
func (e *Event) binSize() int {

	// start with the size of the "size" header + size of metric + size of time
	size := 32

	// get the size of the tags
	size += e.Tags.binSize()

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
	// tagset header
	size := 1
	// key/val headers
	size += len(m) * 2
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
	e.mut.Lock()
	e.wait.Done()
	e.mut.Unlock()
}

// WaitAdd increments ot the event's waitgroup counter
func (e *Event) WaitInc() {
	e.mut.Lock()
	e.wait.Add(1)
	e.mut.Unlock()
}

// EventPasser provides a method for passing an event down a step in the pipeline
type EventPasser interface {
	PassEvent(e *Event)
}

func NewEvent() *Event {
	e := &Event{
		Tags: &TagSet{},
	}
	return e
}

// Get any value on an event as a string
func (e *Event) Get(key string) string {
	if e.Tags == nil {
		return ""
	}
	return e.Tags.Get(key)
}

func (e *Event) IndexName() string {
	return e.Tags.String()
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
