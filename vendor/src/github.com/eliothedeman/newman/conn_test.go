package newman

import (
	"bytes"
	"fmt"
	"testing"
)

func TestNewConn(t *testing.T) {
	in := bytes.NewBuffer(nil)
	c := NewConn(WrapNoopCloser(in), Encrypted)

	if c.options != Encrypted {
		t.Fatal()
	}

	_, err := in.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

}

type testMessage string

func newTestMessage(s string) *testMessage {
	t := testMessage(s)
	return &t
}

func (t *testMessage) MarshalBinary() ([]byte, error) {
	return []byte(*t), nil
}

func (t *testMessage) UnmarshalBinary(buff []byte) error {
	*t = testMessage(buff)
	return nil
}

func TestWriteNext(t *testing.T) {
	in := NewConn(WrapNoopCloser(bytes.NewBuffer(nil)))
	out := NewConn(in.rwc)

	err := in.Write(newTestMessage("hello world"))
	if err != nil {
		t.Fatal(err)
	}

	m := newTestMessage("")

	err = out.Next(m)
	if err != nil {
		t.Fatal(err)
	}

	if *m != "hello world" {
		t.Fatal(m, "hello world")
	}

}

func TestGenerate(t *testing.T) {
	in := NewConn(WrapNoopCloser(bytes.NewBuffer(nil)))
	out := NewConn(in.rwc)

	go func() {
		// write out 1000 messages
		for i := 0; i < 1000; i++ {
			err := in.Write(newTestMessage(fmt.Sprintf("hello world %d", i)))
			if err != nil {
				t.Fatal(err)
			}
		}
	}()

	i := 0
	c, d := out.Generate(func() Message {
		return newTestMessage("")
	})

	for m := range c {
		if string(*m.(*testMessage)) != fmt.Sprintf("hello world %d", i) {
			t.Fatal()
		}
	}

	d()

}
