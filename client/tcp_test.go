package client

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/eliothedeman/bangarang/event"
)

var (
	numEvents = 0
)

type testTcpClient struct {
	c    *TcpClient
	buff *bytes.Buffer
}

func newTestTcpClient() *testTcpClient {
	t := &testTcpClient{}
	c, _ := NewTcpClient("", "json", 4)
	buff := bytes.NewBuffer([]byte{})
	c.conn = buff
	t.c = c
	t.buff = buff
	return t
}

func newTestEvent() *event.Event {
	numEvents += 1
	e := &event.Event{
		Host:    fmt.Sprintf("test_%d", numEvents),
		Service: "test_service",
		Metric:  float64(numEvents),
	}

	return e
}

func TestNewTcpClient(t *testing.T) {
	c := newTestTcpClient()
	e := newTestEvent()
	err := c.c.Send(e)
	if err != nil {
		t.Error(err)
	}

	expected := `{ "host":"test_1","service":"test_service","sub_type":"","metric":1,"occurences":0,"tags":null,"status":0}`
	if expected != c.buff.String()[:c.buff.Len()-len(DELIMITER)] {
		t.Error(c.buff.String())
	}
}
