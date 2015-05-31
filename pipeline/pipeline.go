package pipeline

import (
	"log"
	"runtime"
	"time"

	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/event"
	"github.com/eliothedeman/bangarang/provider"
)

type Pipeline struct {
	keepAliveAge       time.Duration
	keepAliveCheckTime time.Duration
	globalPolicy       *alarm.Policy
	escalations        alarm.AlarmCollection
	policies           []*alarm.Policy
	index              *event.Index
	providers          provider.EventProviderCollection
	encodingPool       *event.EncodingPool
}

func NewPipeline(conf *config.AppConfig) *Pipeline {
	p := &Pipeline{
		encodingPool:       event.NewEncodingPool(event.EncoderFactories[*conf.Encoding], event.DecoderFactories[*conf.Encoding], runtime.NumCPU()),
		keepAliveAge:       conf.KeepAliveAge,
		keepAliveCheckTime: 30 * time.Second,
		escalations:        *conf.Escalations,
		index:              event.NewIndex(conf.DbPath),
		providers:          conf.EventProviders,
		policies:           conf.Policies,
		globalPolicy:       conf.GlobalPolicy,
	}
	return p
}

func (p *Pipeline) checkExpired() {
	for {
		time.Sleep(p.keepAliveCheckTime)

		hosts := p.index.GetExpired(p.keepAliveAge)
		for _, host := range hosts {
			e := &event.Event{
				Host:    host,
				Service: "KeepAlive",
				Metric:  float64(p.keepAliveAge),
			}
			p.Process(e)
		}
	}
}

func (p *Pipeline) Start() {
	go p.checkExpired()
	dst := make(chan *event.Event, 25)

	// start up all of the providers
	for _, ep := range p.providers {
		go ep.Start(dst)
	}

	// fan in all of the providers and process them
	go func() {
		var e *event.Event
		for {
			// recieve the event
			e = <-dst

			// process the event
			p.Process(e)
		}
	}()

}

// Run the given event though the pipeline
func (p *Pipeline) Process(e *event.Event) int {
	if p.index == nil {
		p.index = event.NewIndex("event.db")
	}

	if p.globalPolicy != nil {
		if !p.globalPolicy.CheckMatch(e) || !p.globalPolicy.CheckNotMatch(e) {
			return event.OK
		}
	}

	p.index.PutEvent(e)
	for _, pol := range p.policies {
		if pol.Matches(e) {
			act := pol.Action(e)

			// if there is an action to be taken
			if act != "" {

				// create a new incident for this event
				in := p.NewIncident(pol.Name, e)

				// dedup the incident
				if p.Dedupe(in) {

					// update the incident in the index
					if in.Status != event.OK {
						p.index.PutIncident(in)
					} else {
						p.index.DeleteIncidentById(in.IndexName())
					}

					// fetch the escalation to take
					esc, ok := p.escalations[act]
					if ok {

						// send to every alarm in the escalation
						for _, a := range esc {
							a.Send(in)
						}
					} else {
						log.Println("unknown escalation", act)
					}
				}
			}
		}
	}

	return e.Status
}

// returns true if this is a new incident, false if it is a duplicate
func (p *Pipeline) Dedupe(i *event.Incident) bool {
	old := p.index.GetIncident(i.IndexName())

	if old == nil {
		return i.Status != event.OK
	}

	return old.Status != i.Status
}

func (p *Pipeline) ListIncidents() []*event.Incident {
	return p.index.ListIncidents()
}

func (p *Pipeline) PutIncident(in *event.Incident) {
	if in.Id == 0 {
		in.Id = p.index.GetIncidentCounter()
		p.index.UpdateIncidentCounter(in.Id + 1)
	}
	p.index.PutIncident(in)
}

func (p *Pipeline) NewIncident(policy string, e *event.Event) *event.Incident {
	return event.NewIncident(policy, e)
}
