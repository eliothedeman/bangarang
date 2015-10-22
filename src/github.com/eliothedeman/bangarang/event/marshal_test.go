package event

import "testing"

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

func TestUnmarshal(t *testing.T) {
	e := newTestEvent("machine01.deployment.company.com", "load", 2.001)
	e.Tags = append(e.Tags, KeyVal{Key: "hello", Value: "world"})

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

	for _, v := range e.Tags {
		if n.Tags.Get(v.Key) != v.Value {
			t.Fatalf("wanted: %v got %v", v.Value, n.Tags.Get(v.Key))
		}
	}

}

func BenchmarkMarshalBinary(b *testing.B) {
	e := newTestEvent("machine01.deployment.company.com", "load", 2.001)
	e.Tags = append(e.Tags, KeyVal{Key: "hello", Value: "world"})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.MarshalBinary()
	}
}

func BenchmarkUnmarshalBinary(b *testing.B) {
	e := newTestEvent("machine01.deployment.company.com", "load", 2.001)
	e.Tags = append(e.Tags, KeyVal{Key: "hello", Value: "world"})

	buff, _ := e.MarshalBinary()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.UnmarshalBinary(buff)
	}
}
