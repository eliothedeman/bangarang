package alarm

import (
	"testing"

	"github.com/eliothedeman/bangarang/event"
)

func TestMatchAll(t *testing.T) {
	e := &event.Event{
		Host: "elioasdft.elasdfiothedmen.com",
	}
	p := &Policy{}
	p.Match = map[string]string{
		"Host": "eliot.*",
	}
	p.Compile()

	t.Error(p.MatchAny(e))

}
