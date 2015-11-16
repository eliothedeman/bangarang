package newman

import (
	"encoding/binary"
	"io"
	"time"

	"golang.org/x/net/context"
)

const (
	DefaultBufferSize        = 1024 * 100 // 100k
	Compression       uint64 = 1 << iota
	Encrypted
)

// NoopCloser io.ReadWriteCloser has a noop for "Close"
type NoopCloser struct {
	io.ReadWriter
}

// Close noop
func (n *NoopCloser) Close() error {
	return nil
}

// WrapNoopCloser wraps a io.ReadWriter in a io.ReadWriteCloser that does nothing on "Close"
func WrapNoopCloser(rw io.ReadWriter) io.ReadWriteCloser {
	return &NoopCloser{rw}
}

// Conn provides a wrapper around an io.ReadWriteCloser for sending size delimited data over any reader/writer
type Conn struct {
	buff         []byte
	nextSizeBuff []byte
	rwc          io.ReadWriteCloser
	options      uint64
	ctx          context.Context
}

// NewConn returns a new Conn with the appropriate configuration
func NewConn(rwc io.ReadWriteCloser, options ...uint64) *Conn {
	c := &Conn{
		buff:         make([]byte, DefaultBufferSize),
		nextSizeBuff: make([]byte, 8),
		rwc:          rwc,
		ctx:          context.Background(),
	}

	// assign all options
	for _, o := range options {
		c.options |= o
	}

	return c
}

// Write a message to the connection
func (c *Conn) Write(m Message) error {
	// encode the message as binary
	buff, err := m.MarshalBinary()
	if err != nil {
		return err
	}

	// write out the message on the connection
	return c.writeNextBuffer(buff)
}

// Next reads the next message off the line and unmarshals it into the given message
func (c *Conn) Next(m Message) error {

	// fill the message buffer
	size, err := c.readNextBuffer()
	if err != nil {
		return err
	}

	// unmarshal the message
	return m.UnmarshalBinary(c.buff[:size])
}

// Generate returns a channel that will recieve messages as they come in off the line
// Only safe to run once per Conn
func (c *Conn) Generate(f func() Message) (<-chan Message, context.CancelFunc) {

	// output channel
	out := make(chan Message, 10)

	// set up the context to cancel the generation
	ctx, stop := context.WithCancel(c.ctx)
	c.ctx = ctx

	// read messages as they come off the line and send them on the output channel
	go func() {
		var err error
		var m Message

		// loop until something bad happens
		for {
			select {

			// check to see if the generation should be stopped
			case <-c.ctx.Done():
				close(out)
				return

			default:
				// make a new message
				m = f()

				// read the next message off the line
				err = c.Next(m)
				if err != nil {
					// close the out put channel and return
					close(out)
					return
				}

				out <- m
			}
		}
	}()

	return out, stop
}

// writeNextBuffer writes the next message legnth and message to the enclosed writer
func (c *Conn) writeNextBuffer(buff []byte) error {
	// write out the size of the next message
	binary.LittleEndian.PutUint64(c.nextSizeBuff, uint64(len(buff)))
	err := writeIntoBuffer(c.nextSizeBuff, c.rwc)
	if err != nil {
		return err
	}

	// write the next message
	return writeIntoBuffer(buff, c.rwc)
}

// readNextBuffer reads the next message into the message buffer
func (c *Conn) readNextBuffer() (int, error) {
	// read the size of the next message
	err := readIntoBuffer(c.nextSizeBuff, c.rwc)
	if err != nil {
		return 0, err
	}

	// find the size of the next message
	size := int(binary.LittleEndian.Uint64(c.nextSizeBuff))

	// make sure the buffer is large enough to hold the next message
	if size > len(c.buff) {

		// make a new buffer twice as large as we need to avoid reallocating
		c.buff = make([]byte, size*2)
	}

	// read the next message into the buffer
	return size, readIntoBuffer(c.buff[:size], c.rwc)
}

// writeIntoBuffer writes an entire message to the io.Writer
func writeIntoBuffer(buff []byte, w io.Writer) error {

	// this will be used to back off the connection if the connection's buffers are full
	backOffTime := time.Millisecond

	x := len(buff)
	var n, l int
	var err error
	// read until the buffer size is met
	for l < x {
		n, err = w.Write(buff[l:])
		if err != nil {
			return err
		}

		// back off if nothing could be written
		if n == 0 {
			time.Sleep(backOffTime)
			backOffTime = backOffTime * 2

		} else {
			backOffTime = time.Millisecond

			// add the written amount to the write counter
			l += n
		}

	}

	return nil

}

// readIntoBuffer will read off of a rwc until it the buffer has been filled
func readIntoBuffer(buff []byte, r io.Reader) error {

	// this will be used to back off the connection when there is nothing to read
	backOffTime := time.Millisecond

	x := len(buff)
	var n, l int
	var err error
	// read until the buffer size is met
	for l < x {
		n, err = r.Read(buff[l:])
		if err != nil {
			return err
		}

		// back off if we got nothing
		if n == 0 {
			time.Sleep(backOffTime)
			backOffTime = backOffTime * 2
		} else {
			// reset the backOffTime
			backOffTime = time.Millisecond

			// add the read amount to the length counter
			l += n
		}

	}

	return nil
}
