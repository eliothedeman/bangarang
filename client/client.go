package client

import "github.com/eliothedeman/bangarang/event"

type Client interface {
	Send(e *event.Event) error
}
