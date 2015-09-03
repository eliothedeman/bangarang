package event

import "testing"

func TestMarshalBinary(t *testing.T) {
	e := &Event{
		Host:    "hello",
		Service: "what's up",
	}
	buff, err := e.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	if len(buff) != e.binSize() {
		t.Fail()
	}
}

func TestBinSize(t *testing.T) {

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
