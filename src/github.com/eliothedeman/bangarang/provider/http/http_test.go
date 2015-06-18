package http

import (
	"testing"

	"github.com/eliothedeman/bangarang/provider"
)

func TestNewHttp(t *testing.T) {
	h := NewHTTPProvider()
	if h == nil {
		t.Fail()
	}

	if _, ok := h.(provider.EventProvider); !ok {
		t.Error("HTTPProvider does not impliement provider.EventProvider")
	}
}
