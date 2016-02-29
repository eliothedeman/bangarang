package client

import (
	"fmt"
	"net"

	"github.com/eliothedeman/bangarang/event"
	"github.com/eliothedeman/newman"
)

// A TCPClient that provides methods for inserting events into bangarang using the tcp provider
type TCPClient struct {
	c    *newman.Conn
	Host string
	Port int
}

// Dial attempts to connect to the bangarang instance via tcp
func (t *TCPClient) Dial() error {

	// attemp to connect to the server
	// TODO add support for ipv6
	raddr, err := net.ResolveTCPAddr("tcp4", fmtHost(t.Host, t.Port))
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp4", nil, raddr)
	if err != nil {
		return err
	}

	t.c = newman.NewConn(conn)
	return nil
}

// Send a single event to the server
func (t *TCPClient) Send(e *event.Event) error {
	return t.c.Write(e)
}

// BatchSend sends a slice of events in serial
func (t *TCPClient) BatchSend(e []*event.Event) error {
	var err error
	for _, x := range e {
		err = t.Send(x)
		if err != nil {
			return err
		}
	}

	return err
}

func fmtHost(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}
