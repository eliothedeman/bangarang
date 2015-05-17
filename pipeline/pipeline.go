package pipeline

import (
	"log"

	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/event"
)

type Pipeline struct {
	port        int
	escalations []*alarm.Escalation
	index       *event.Index
}

func (p *Pipeline) Process(e *event.Event) int {
	if p.index == nil {
		p.index = event.NewIndex()
	}

	p.index.Put(e)
	for _, v := range p.escalations {
		if v.Match(e) {
			if v.StatusOf(e) != event.OK && e.StatusChanged() {
				for _, a := range v.Alarms {
					err := a.Send(e)
					if err != nil {
						log.Println(err)
					}
				}
				return e.Status
			}
		}
	}
	return e.Status
}
