package event

import (
	"bytes"
	"errors"
	"sort"
)

// KeyVal is a key value pair, used in matching to tags on events
type KeyVal struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// TagSet is a collection of key/val pairs that will always give the same order of the tags
type TagSet []KeyVal

func NewTagset(size int) *TagSet {
	ts := make(TagSet, 0, size)
	return &ts
}

func (t *TagSet) Len() int {
	if t == nil {
		return 0
	}

	return len(*t)
}

func (t *TagSet) ForEach(f func(k, v string)) {
	if t == nil {
		return
	}
	for _, tag := range *t {
		f(tag.Key, tag.Value)
	}
}

// Get returns the value of a key in linear time. An empty string if it wasn't found
func (t *TagSet) Get(key string) string {
	if t == nil {
		return ""
	}
	for _, tag := range *t {
		if tag.Key == key {
			return tag.Value
		}
	}
	return ""
}

func (t *TagSet) Set(key, val string) {
	if t == nil {
		return
	}
	*t = append(*t, KeyVal{Key: key, Value: val})
}

func (t *TagSet) MarshalBinary(buff []byte) error {
	if t == nil {
		errors.New("attempted to encode a nil tagset")
	}

	if len(buff) < t.binSize() {
		return errors.New("unable to binary marshal tagset, given buffer is too small")
	}

	tmp := 0
	offset := 0
	for _, v := range *t {
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

func (t *TagSet) String() string {
	if t == nil {
		return ""
	}

	// make sure it is sorted, so we can give the same result every time
	t.SortByKey()

	// make a new buffer that will be able to hold the entire tagset
	buff := bytes.NewBuffer(nil)
	t.ForEach(func(k, v string) {
		buff.WriteString(k)
		buff.WriteString(":")
		buff.WriteString(v)
		buff.WriteString(",")

	})
	return buff.String()
}

func (t *TagSet) binSize() int {
	if t == nil {
		return 0
	}
	// tagset header
	size := 1
	// key/val headers
	size += len(*t) * 2
	for _, tag := range *t {

		// size of key
		size += sizeOfString(tag.Key)

		// size of value
		size += sizeOfString(tag.Value)
	}
	return size

}

type by func(k1, k2 KeyVal) bool

type tagSetSorter struct {
	ts *TagSet
	by by
}

func (t *tagSetSorter) Len() int {
	return len(*t.ts)
}

func (t *tagSetSorter) Less(i, j int) bool {
	ts := *t.ts
	return t.by(ts[i], ts[j])
}

func (t *tagSetSorter) Swap(i, j int) {
	ts := *t.ts
	ts[i], ts[j] = ts[j], ts[i]
}

func (t *TagSet) SortByKey() {
	tss := &tagSetSorter{
		ts: t,
		by: func(k1, k2 KeyVal) bool {
			return k1.Key < k2.Key
		},
	}
	sort.Sort(tss)
}

func (t *TagSet) SortByValue() {
	tss := &tagSetSorter{
		ts: t,
		by: func(k1, k2 KeyVal) bool {
			return k1.Value < k2.Value
		},
	}
	sort.Sort(tss)
}
