package tcp

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/eliothedeman/bangarang/event"
	"github.com/eliothedeman/bangarang/provider"
	"github.com/eliothedeman/newman"
)

func setup() {
	event.INDEX_FILE_NAME = fmt.Sprintf("/tmp/bangarang-index-%d.db", num)
}

var num = 1000

func newTestTCP() (provider.EventProvider, int) {
	num += 1
	setup()
	p := NewTCPProvider()
	conf := p.ConfigStruct().(*TCPConfig)
	listen := fmt.Sprintf("0.0.0.0:%d", 9099+num)
	conf.Listen = listen
	err := p.Init(conf)
	if err != nil {
		log.Fatal(err)
	}
	return p, 9099 + num
}

func randomString(l int) string {
	buff := make([]byte, l)
	for i := range buff {
		buff[i] = uint8(65 + (rand.Uint32() % 26))
	}

	return string(buff)
}

func newTestEvent() *event.Event {
	e := event.NewEvent()
	e.Tags.Set("host", randomString(rand.Int()%50))
	e.Tags.Set("service", randomString(rand.Int()%50))
	e.Tags.Set("sub_service", randomString(rand.Int()%50))
	e.Time = time.Now()
	e.Metric = rand.Float64() * 100
	return e
}

type testPasser struct {
	in chan *event.Event
}

func (t *testPasser) Pass(e *event.Event) {
	t.in <- e
}

func TestSendSingle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping because of short test")
	}
	p, port := newTestTCP()

	tp := &testPasser{
		in: make(chan *event.Event),
	}

	p.Start(tp)

	c, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatal(err)
	}

	e := newTestEvent()

	conn := newman.NewConn(c)
	conn.Write(e)

	ne := <-tp.in

	ne.Tags.ForEach(func(k, v string) {
		if e.Get(k) != v {
			t.Fatalf("Wanted %s got %s for %s", e.Get(k), v, k)
		}
	})

	if ne.Metric != e.Metric {
		t.Fatal()
	}
}

func TestMany(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping because of short test")
	}
	p, port := newTestTCP()
	tp := &testPasser{
		in: make(chan *event.Event),
	}

	p.Start(tp)

	c, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatal(err)
	}

	conn := newman.NewConn(c)

	go func() {
		// send 10000 events
		for i := 0; i < 10000; i++ {
			e := newTestEvent()
			conn.Write(e)
		}

	}()

	for i := 0; i < 10000; i++ {
		<-tp.in
	}
}
