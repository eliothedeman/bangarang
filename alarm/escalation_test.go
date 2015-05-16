package alarm

import (
	"testing"

	"github.com/eliothedeman/bangarang/event"
)

func TestMatchAny(t *testing.T) {
	e := &event.Event{
		Host: "eliot.com",
	}
	p := &Policy{}
	p.Match = map[string]string{
		"host": "eliot.*",
	}
	p.NotMatch = map[string]string{
		"sub_service": "testing",
	}
	p.Compile()

	if !p.MatchAny(e) {
		t.Fail()
	}

	e.SubService = "testing"

	if p.MatchAny(e) {
		t.Fail()
	}

}
