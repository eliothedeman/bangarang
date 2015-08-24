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

const (
	KEEP_ALIVE_SERVICE_NAME = "KeepAlive"
)

type Pipeline struct {
	keepAliveAge       time.Duration
	keepAliveCheckTime time.Duration
	globalPolicy       *alarm.Policy
	escalations        *alarm.Collection
	policies           map[string]*alarm.Policy
	index              *event.Index
	providers          provider.EventProviderCollection
	encodingPool       *event.EncodingPool
	config             *config.AppConfig
	tracker            *Tracker
	pauseCache         map[*event.Event]struct{}
	unpauseChan        chan struct{}
	in                 chan *event.Event
}

// Passer provides a method for passing an event down a step in the pipeline
type Passer interface {
	Pass(e event.Event)
}

func NewPipeline(conf *config.AppConfig) *Pipeline {
	p := &Pipeline{
		encodingPool:       event.NewEncodingPool(event.EncoderFactories[conf.Encoding], event.DecoderFactories[conf.Encoding], runtime.NumCPU()),
		keepAliveAge:       conf.KeepAliveAge,
		keepAliveCheckTime: 10 * time.Second,
		in:                 make(chan *event.Event),
		unpauseChan:        make(chan struct{}),
		tracker:            NewTracker(),
		escalations:        &alarm.Collection{},
		index:              event.NewIndex(),
	}

	p.Refresh(conf)

	logrus.Debug("Starting expiration checker")
	go p.checkExpired()

	return p
}

func (p *Pipeline) Pass(e event.Event) {
	p.in <- &e
}

// refresh load all config params that don't require a restart
func (p *Pipeline) Refresh(conf *config.AppConfig) {
	p.pause()

	// if the config has changed at all, refresh the index
	if p.config == nil || string(conf.Hash) != string(p.config.Hash) {
		go p.tracker.Start()
	}

	p.escalations = &conf.Escalations

	if conf.EventProviders != nil {
		p.providers = *conf.EventProviders
	}

	p.policies = conf.Policies
	p.keepAliveAge = conf.KeepAliveAge
	p.globalPolicy = conf.GlobalPolicy

	if p.globalPolicy != nil {
		p.globalPolicy.Compile()
	}

	// update to the new config
	p.config = conf
	p.unpause()

	// start up all of the providers
	logrus.Infof("Starting %d providers", len(p.providers.Collection))
	for name, ep := range p.providers.Collection {
		logrus.Infof("Starting event provider %s", name)
		go ep.Start(p)
	}
}

// unpause resume processing jobs
func (p *Pipeline) unpause() {
	logrus.Info("Unpausing pipeline")
	p.unpauseChan <- struct{}{}
	<-p.unpauseChan
}

// pause stop processing events
func (p *Pipeline) pause() {
	logrus.Info("Pausing pipeline")

	// cache the old injest channel
	old := p.in

	// make a temporary channel to catch incomming events
	p.in = nil

	// make a channel to signal the end of the pause
	done := make(chan struct{})
	p.unpauseChan = done

	// start a new goroutine to catch the incomming events
	go func() {
		// make a map to cache the incomming events
		cache := make(map[*event.Event]struct{})
		var e *event.Event
		for {
			select {

			// start caching the events as they come in
			case e = <-old:
				logrus.Debugf("Caching event during pause %+v", e)
				cache[&e] = struct{}{}

			// when the pause is complete, revert to the old injestion channel
			case <-done:

				// set the cached event channel
				p.in = old

				// restart the pipeline
				p.Start()

				// empty the cache
				for e, _ := range cache {
					logrus.Debugf("Proccessing cached event after unpause %+v", *e)
					old <- e
				}

				// signal the unpause function that we are done with the unpause
				p.unpauseChan <- struct{}{}
				return
			}
		}
	}()
}

func (p *Pipeline) GetTracker() *Tracker {
	return p.tracker
}

func (p *Pipeline) GetConfig() *config.AppConfig {
	return p.config
}

func (p *Pipeline) checkExpired() {
	var events []*event.Event
	for {
		time.Sleep(p.keepAliveCheckTime)
		logrus.Info("Checking for expired events.")

		// get keepalive events for all known hosts
		events = createKeepAliveEvents(p.tracker.HostTimes())
		logrus.Infof("Found %d hosts with keepalives", len(events))

		// process every event as if it was an incomming event
		for _, e := range events {
			p.Process(e)
		}
	}
}

// create keep alive events for each hostname -> time pair
func createKeepAliveEvents(times map[string]time.Time) []*event.Event {
	n := time.Now()
	events := make([]*event.Event, len(times))
	x := 0
	for host, t := range times {
		events[x] = &event.Event{
			Host:    host,
			Metric:  n.Sub(t).Seconds(),
			Service: KEEP_ALIVE_SERVICE_NAME,
		}
		x += 1
	}

	return events
}

func (p *Pipeline) Start() {
	logrus.Info("Starting pipeline")

	// fan in all of the providers and process them
	go func() {
		var e *event
		for {

			// if the injest channel is nil, stop
			if p.in == nil {
				logrus.Info("Pipeline is paused, stopping pipeline")
				return
			}

			// recieve the event
			e = <-p.in

			logrus.Debugf("Beginning processing %+v", e)
			// process the event
			p.Process(e)
			logrus.Debugf("Done processing %+v", e)
		}
	}()
}

func (p *Pipeline) ProcessIncident(in *event.Incident) {
	// dedup the incident
	if p.Dedupe(in) {

		// update the incident in the index
		if in.Status != event.OK {
			p.index.PutIncident(in)
		} else {
			p.index.DeleteIncidentById(in.IndexName())
		}

		// fetch the escalation to take
		esc, ok := p.escalations.Collection()[in.Escalation]
		if ok {

			// send to every alarm in the escalation
			for _, a := range esc {
				a.Send(in)
			}
		} else {
			log.Println("unknown escalation", in.Escalation)
		}
	}
}

// Run the given event though the pipeline
func (p *Pipeline) Process(e *event.Event) int {
	if p.globalPolicy != nil {
		if !p.globalPolicy.CheckMatch(e) || !p.globalPolicy.CheckNotMatch(e) {
			return event.OK
		}
	}

	// track stas for this event
	p.tracker.TrackEvent(e)

	// process this event on every policy
	for _, pol := range p.policies {
		pol.Process(*e, func(in *event.Incident) {
			e.Wait.Add(1)
			p.ProcessIncident(in)
		})
	}

	return e.Status
}

func (p *Pipeline) GetIndex() *event.Index {
	return p.index
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

func (p *Pipeline) NewIncident(policy string, escalation string, e *event.Event) *event.Incident {
	return event.NewIncident(policy, escalation, e)
}
