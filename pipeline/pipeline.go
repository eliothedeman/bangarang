package pipeline

import (
	"log"

	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/alarm/console"
	"github.com/eliothedeman/bangarang/event"
)

type Pipeline struct {
	port        int
	escalations []*alarm.Escalation
}

func newTestPipeline() *Pipeline {
	return &Pipeline{
		escalations: []*alarm.Escalation{
			&alarm.Escalation{
				Policy: alarm.Policy{
					Match:    make(map[string]string),
					NotMatch: make(map[string]string),
				},
				Alarms: []alarm.Alarm{
					console.NewConsole(),
				},
			},
		},
	}
}

func (p *Pipeline) Process(e *event.Event) {
	for _, v := range p.escalations {
		if v.Match(e) {
			if v.StatusOf(e) != event.OK {
				for _, a := range v.Alarms {
					err := a.Send(e)
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}
}
