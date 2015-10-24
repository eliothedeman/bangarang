package event

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"sort"
	"sync"
	"time"
	"unsafe"
)

const (
	OK = iota
	WARNING
	CRITICAL
)

// KeyVal is a key value pair, used in matching to tags on events
type KeyVal struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// TagSet is a collection of key/val pairs that will always give the same order of the tags
type TagSet []KeyVal

func (t TagSet) ForEach(f func(k, v string)) {
	for _, tag := range t {
		f(tag.Key, tag.Value)
	}
}

// Get returns the value of a key in linear time. An empty string if it wasn't found
func (t TagSet) Get(key string) string {
	for _, tag := range t {
		if tag.Key == key {
			return tag.Value
		}
	}
	return ""
}

func (t TagSet) Set(key, val string) {
	t = append(t, KeyVal{Key: key, Value: val})
}

func (t TagSet) MarshalBinary(buff []byte) error {
	tmp := 0
	offset := 0
	for _, v := range t {
		tmp = sizeOfString(v.Key)
		buff[offset] = uint8(tmp)
		offset += 1
		copy(buff[offset:offset+tmp], v.Key)
		offset += tmp

		tmp = sizeOfString(v.Value)
		buff[offset] = uint8(tmp)
		offset += 1
		copy(buff[offset:offset+tmp], v.Value)
		offset += tmp
	}

	return nil
}

func (t TagSet) String() string {
	t.SortByKey()
	buff := make([]byte, t.binSize())
	t.MarshalBinary(buff)
	return string(buff)
}

func (t TagSet) binSize() int {
	// tagset header
	size := 1
	// key/val headers
	size += len(t) * 2
	for _, tag := range t {

		// size of key
		size += sizeOfString(tag.Key)

		// size of value
		size += sizeOfString(tag.Value)
	}
	return size

}

type by func(k1, k2 KeyVal) bool

type tagSetSorter struct {
	ts TagSet
	by by
}

func (t *tagSetSorter) Len() int {
	return len(t.ts)
}

func (t *tagSetSorter) Less(i, j int) bool {
	return t.by(t.ts[i], t.ts[j])
}

func (t *tagSetSorter) Swap(i, j int) {
	t.ts[i], t.ts[j] = t.ts[j], t.ts[i]
}

func (t TagSet) SortByKey() {
	tss := &tagSetSorter{
		ts: t,
		by: func(k1, k2 KeyVal) bool {
			return k1.Key < k2.Key
		},
	}
	sort.Sort(tss)
}

func (t TagSet) SortByValue() {
	tss := &tagSetSorter{
		ts: t,
		by: func(k1, k2 KeyVal) bool {
			return k1.Value < k2.Value
		},
	}
	sort.Sort(tss)
}

//go:generate ffjson $GOFILE
//go:generate msgp $GOFILE

// Event represents a metric as it passes through the pipeline.
// It holds meta data about the metric, as well as methods to trace the event as it is processed
type Event struct {
	Metric    float64   `json:"metric" msg:"metric"`
	Tags      TagSet    `json:"tags" msg:"tags"`
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
	e.Tags = make(TagSet, int(buff[offset]))
	offset += 1
	i = 0
	kv := KeyVal{}
	for offset < len(buff) {

		// key
		l = int(buff[offset])
		offset += 1
		kv.Key = string(buff[offset : offset+l])
		offset += l

		l = int(buff[offset])
		offset += 1
		kv.Value = string(buff[offset : offset+l])
		offset += l

		e.Tags[i] = kv
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

	buff[offset] = uint8(len(e.Tags))
	offset += 1

	// tags
	for _, v := range e.Tags {
		tmp = sizeOfString(v.Key)
		buff[offset] = uint8(tmp)
		offset += 1
		copy(buff[offset:offset+tmp], v.Key)
		offset += tmp

		tmp = sizeOfString(v.Value)
		buff[offset] = uint8(tmp)
		offset += 1
		copy(buff[offset:offset+tmp], v.Value)
		offset += tmp
	}

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
