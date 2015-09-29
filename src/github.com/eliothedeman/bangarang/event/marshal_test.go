package event

import "testing"

func TestMarshalBinary(t *testing.T) {
	e := &Event{
		Host:    "hello",
		Service: "what's up",
		Metric:  3.0005,
	}
	buff, err := e.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	if len(buff) != e.binSize() {
		t.Fail()
	}
}

func TestUnmarshal(t *testing.T) {
	e := &Event{
		Host:    "machine01.deployment.company.com",
		Service: "load",
		Metric:  2.001,
		Tags: map[string]string{
			"key": "value",
		},
	}

	buff, _ := e.MarshalBinary()

	n := &Event{}
	err := n.UnmarshalBinary(buff)
	if err != nil {
		t.Fatal(err)
	}

	if e.Metric != n.Metric {
		t.Fatalf("wanted: %v got %v", e.Metric, n.Metric)

	}

	if e.Host != n.Host {
		t.Fatalf("wanted: %v got %v", e.Host, n.Host)
	}

	if e.Service != n.Service {
		t.Fatalf("wanted: %v got %v", e.Service, n.Service)
	}

	if e.SubService != n.SubService {
		t.Fatalf("wanted: %v got %v", e.SubService, n.SubService)
	}

	for k, v := range e.Tags {
		if n.Tags[k] != v {
			t.Fatalf("wanted: %v got %v", v, n.Tags[k])
		}
	}

}

func BenchmarkMarshalBinary(b *testing.B) {
	e := &Event{
		Host:    "hello",
		Service: "what's up",
		Tags: map[string]string{
			"key": "value",
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.MarshalBinary()
	}
}

func BenchmarkUnmarshalBinary(b *testing.B) {
	e := &Event{
		Host:    "hello",
		Service: "what's up",
		Tags: map[string]string{
			"key": "value",
		},
	}

	buff, _ := e.MarshalBinary()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.UnmarshalBinary(buff)
	}
}
