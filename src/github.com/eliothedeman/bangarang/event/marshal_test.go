package event

import (
	"fmt"
	"testing"
)

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

	fmt.Printf("%+v\n", e)
	fmt.Printf("%+v\n", n)

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
