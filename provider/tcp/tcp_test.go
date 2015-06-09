package tcp

import (
	"net"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/client"
	"github.com/eliothedeman/bangarang/event"
)

func TestNewTCPProvider(t *testing.T) {
	p := NewTCPProvider()
	if p == nil {
		t.Fail()
	}
}

func TestConfig(t *testing.T) {
	p := NewTCPProvider()
	conf := p.ConfigStruct().(*TCPConfig)
	conf.Encoding = event.ENCODING_TYPE_JSON
	conf.Listen = ":8083"
	conf.MaxDecoders = 4

	err := p.Init(conf)
	if err != nil {
		t.Error(err)
	}
}

func TestBadAddr(t *testing.T) {
	p := NewTCPProvider()
	conf := p.ConfigStruct().(*TCPConfig)
	conf.Listen = "10.0.0.1:8081"
	conf.MaxDecoders = 4

	err := p.Init(conf)
	if _, ok := err.(*net.OpError); !ok {
		t.Fatalf("Expecting bad listening address")
	}
}

func TestStart(t *testing.T) {
	p := NewTCPProvider()
	conf := p.ConfigStruct().(*TCPConfig)
	conf.Encoding = event.ENCODING_TYPE_JSON
	conf.Listen = ":8082"
	conf.MaxDecoders = 4
	p.Init(conf)

	cli, err := client.NewTcpClient("localhost:8082", event.ENCODING_TYPE_JSON, 1)
	if err != nil {
		t.Error(err)
	}
	e := &event.Event{}
	e.Host = "this is a test"

	dst := make(chan *event.Event)
	go p.Start(dst)

	go func() {
		err = cli.Send(e)
		if err != nil {
			t.Error(err)
		}
	}()

	e2 := <-dst
	if e2.Host != e.Host {
		t.Fail()
	}
}

func TestMultipleEvents(t *testing.T) {
	p := NewTCPProvider()
	conf := p.ConfigStruct().(*TCPConfig)
	conf.Encoding = event.ENCODING_TYPE_MSGPACK
	conf.Listen = ":8084"
	conf.MaxDecoders = 4
	p.Init(conf)

	cli, err := client.NewTcpClient("localhost:8084", event.ENCODING_TYPE_MSGPACK, 1)
	if err != nil {
		t.Error(err)
	}
	e := &event.Event{}
	e.Host = "this is a test"

	dst := make(chan *event.Event, 10)
	go p.Start(dst)

	go func() {

		for i := 0; i < 5; i++ {
			e.Metric = float64(i)
			err = cli.Send(e)
			if err != nil {
				t.Error(err)
			}
		}
	}()

	for i := 0; i < 5; i++ {
		e2 := <-dst
		logrus.Info(e2)
	}

}
