package client

import (
	"io"
	"net"

	"github.com/eliothedeman/bangarang/event"
)

var (

	// an impossible string of bytes given all available encodings
	DELIMITER = []byte{0, 0, 0, 0, 0, 0, 0, 0}
)

// A client which maintains an open tcp connection to the server
type TcpClient struct {
	rAddr   string
	encoder *event.EncodingPool
	conn    io.Writer
}

// Create and return a new tcp client with it's tcp connection initilized
func NewTcpClient(srvAddr, encoding string, maxEncoders int) (*TcpClient, error) {
	c := &TcpClient{
		rAddr:   srvAddr,
		encoder: event.NewEncodingPool(event.EncoderFactories[encoding], event.DecoderFactories[encoding], maxEncoders),
	}

	return c, c.dial()
}

// establish a tcp connection with the remote server
func (t *TcpClient) dial() error {
	conn, err := net.Dial("tcp", t.rAddr)
	t.conn = conn
	return err
}

// Send the given event over the client's tcp connection
func (t *TcpClient) Send(e *event.Event) error {
	// encode the event
	var buff []byte
	var err error
	t.encoder.Encode(func(enc event.Encoder) {
		buff, err = enc.Encode(e)
	})

	buff = append(buff, DELIMITER...)

	if err != nil {
		return err
	}

	// attempt to send the encoded event
	_, err = t.conn.Write(buff)

	return err
}