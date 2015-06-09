package pipeline

import (
	"log"
	"runtime"
	"time"

	"github.com/Sirupsen/logrus"
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
	config             *config.AppConfig
	tracker            *Tracker
}

func NewPipeline(conf *config.AppConfig) *Pipeline {
	p := &Pipeline{
		encodingPool:       event.NewEncodingPool(event.EncoderFactories[conf.Encoding], event.DecoderFactories[conf.Encoding], runtime.NumCPU()),
		keepAliveAge:       conf.KeepAliveAge,
		keepAliveCheckTime: 30 * time.Second,
		escalations:        *conf.Escalations,
		index:              event.NewIndex(conf.DbPath),
		providers:          *conf.EventProviders,
		policies:           conf.Policies,
		globalPolicy:       conf.GlobalPolicy,
		config:             conf,
		tracker:            NewTracker(),
	}

	go p.tracker.Start()
	return p
}

func (p *Pipeline) GetTracker() *Tracker {
	return p.tracker
}

func (p *Pipeline) GetConfig() *config.AppConfig {
	return p.config
}

func (p *Pipeline) checkExpired() {
	for {
		logrus.Info("Checking for expired events.")
		time.Sleep(p.keepAliveCheckTime)

		// get keepalive events for all known hosts
		events := p.index.GetKeepAlives()
		logrus.Infof("Found %d hosts with keepalives", len(events))

		// process every event as if it was an incomming event
		for _, e := range events {
			p.Process(e)
		}
	}
}

func (p *Pipeline) Start() {

	logrus.Debug("Starting expiration checker")
	go p.checkExpired()
	dst := make(chan *event.Event, 25)

	// start up all of the providers
	logrus.Info("Starting %d providers", len(p.providers))
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
	if p.globalPolicy != nil {
		if !p.globalPolicy.CheckMatch(e) || !p.globalPolicy.CheckNotMatch(e) {
			return event.OK
		}
	}

	p.index.PutEvent(e)

	// track stas for this event
	p.tracker.TrackEvent(e)

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
