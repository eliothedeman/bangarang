package alarm

import "github.com/eliothedeman/bangarang/event"

type Condition struct {
	Greater    *float64 `greater`
	Less       *float64 `less`
	Exactly    *float64 `exactly`
	Occurences int      `occurences`
}

// check if an event satisfies a condition
func (c *Condition) Satisfies(e *event.Event) bool {
	if c.Greater != nil {
		if e.Metric > *c.Greater {
			return true
		}
	}

	if c.Less != nil {
		if e.Metric < *c.Less {
			return true
		}
	}

	if c.Exactly != nil {
		if e.Metric == *c.Exactly {
			return true
		}
	}

	return false
}
