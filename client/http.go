package client

import (
	"bytes"
	"net/http"

	"github.com/eliothedeman/bangarang/event"
)

// A client for bangarang that connects to the server via http calls
type HttpClient struct {
	host    string
	encoder *event.EncodingPool
}

// Create and return a new HttpClient
func NewHttpClient(host, encoding string, maxEncoders int) (*HttpClient, error) {
	c := &HttpClient{
		host:    host,
		encoder: event.NewEncodingPool(event.EncoderFactories[encoding], event.DecoderFactories[encoding], maxEncoders),
	}
	return c, nil
}

// Send the event using an http call
func (h *HttpClient) Send(e *event.Event) error {
	// encode the event
	var buff []byte
	var err error
	h.encoder.Encode(func(enc event.Encoder) {
		buff, err = enc.Encode(e)
	})

	if err != nil {
		return err
	}

	// post the encoded event to the server
	snd_buf := bytes.NewBuffer(buff)
	_, err = http.Post(h.host, "application/json", snd_buf)

	return err
}
