package event

import (
	"math/rand"
	"testing"
	"time"
)
import "github.com/eliothedeman/randutil"

func TestMarshalBinary(t *testing.T) {
	e := newTestEvent("localhost", "what's up", 3.0005)
	buff, err := e.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	if len(buff) != e.binSize() {
		t.Fail()
	}
}

func TestMarshalTagTooBig(t *testing.T) {
	e := NewEvent()
	e.Tags.Set("host", randutil.AlphaString(300))
	buff, err := e.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	ne := NewEvent()
	err = ne.UnmarshalBinary(buff)
	if err != nil {
		t.Fatal(err)
	}

	if len(ne.Tags.Get("host")) != 255 {
		t.Fatal(ne.Tags.Get("host"))
	}

}

func TestMarshalMany(t *testing.T) {
	nte := func() *Event {
		e := NewEvent()
		for i := 0; i < randutil.IntRange(0, 10); i++ {
			e.Tags.Set(randutil.String(randutil.IntRange(1, 10), randutil.Alphanumeric), randutil.String(randutil.IntRange(1, 10), randutil.Alphanumeric))
		}
		e.Time = time.Now()

		e.Metric = rand.Float64()
		return e
	}

	for i := 0; i < 100000; i++ {
		e := nte()
		buff, err := e.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		ne := NewEvent()
		err = ne.UnmarshalBinary(buff)
		if err != nil {
			t.Fatal(err)
		}

		if ne.Tags.String() != e.Tags.String() {
			t.Fatal(e.Tags.String(), ne.Tags.String())
		}

	}

}

func TestUnmarshal(t *testing.T) {
	e := newTestEvent("machine01.deployment.company.com", "load", 2.001)
	e.Tags.Set("hello", "world")

	buff, _ := e.MarshalBinary()

	n := &Event{}
	err := n.UnmarshalBinary(buff)
	if err != nil {
		t.Fatal(err)
	}

	if e.Metric != n.Metric {
		t.Fatalf("wanted: %v got %v", e.Metric, n.Metric)
	}

	if e.Time.UnixNano() != n.Time.UnixNano() {
		t.Fatalf("wanted: %v got %v", e.Time, n.Time)
	}
	e.Tags.ForEach(func(k, v string) {
		if n.Tags.Get(k) != v {
			t.Fatalf("wanted: %v got %v", v, n.Tags.Get(k))
		}

	})

}

func BenchmarkMarshalBinary(b *testing.B) {
	e := newTestEvent("machine01.deployment.company.com", "load", 2.001)
	e.Tags.Set("hello", "world")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.MarshalBinary()
	}
}

func BenchmarkUnmarshalBinary(b *testing.B) {
	e := newTestEvent("machine01.deployment.company.com", "load", 2.001)
	e.Tags.Set("hello", "world")

	buff, _ := e.MarshalBinary()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.UnmarshalBinary(buff)
	}
}
