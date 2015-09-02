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

// Pipeline
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
	pauseChan          chan struct{}
	unpauseChan        chan struct{}
	in                 chan *event.Event
}

// NewPipeline
func NewPipeline(conf *config.AppConfig) *Pipeline {
	p := &Pipeline{
		encodingPool:       event.NewEncodingPool(event.EncoderFactories[conf.Encoding], event.DecoderFactories[conf.Encoding], runtime.NumCPU()),
		keepAliveAge:       conf.KeepAliveAge,
		keepAliveCheckTime: 10 * time.Second,
		in:                 make(chan *event.Event),
		unpauseChan:        make(chan struct{}),
		pauseChan:          make(chan struct{}),
		tracker:            NewTracker(),
		escalations:        &alarm.Collection{},
		index:              event.NewIndex(),
	}
	p.Start()

	p.Refresh(conf)

	logrus.Debug("Starting expiration checker")
	go p.checkExpired()

	return p
}

func (p *Pipeline) Pass(e *event.Event) {
	p.in <- e
}

// only adds polcies that are not already known of
func (p *Pipeline) refreshPolicies(m map[string]*alarm.Policy) {
	if p.policies == nil {
		p.policies = make(map[string]*alarm.Policy)
	}
	for k, v := range m {

		// if the name of the new polcy is not known of, insert it
		if _, inMap := p.policies[k]; !inMap {
			p.policies[k] = v
		} else {

			// stop the policy if not. Stops the memory leak
			if p.policies[k] != v {
				v.Stop()
			}
		}
	}
}

// RemovePolicy will stop and remove the policy if it exists
func (p *Pipeline) RemovePolicy(name string) {
	p.Pause()
	pol, ok := p.policies[name]
	if ok {
		log.Println("stopping", name)
		pol.Stop()
		delete(p.policies, name)
	}
	p.Unpause()
}

// refresh load all config params that don't require a restart
func (p *Pipeline) Refresh(conf *config.AppConfig) {
	p.Pause()

	// if the config has changed at all, refresh the index
	if p.config == nil || string(conf.Hash) != string(p.config.Hash) {
		if !p.tracker.Started() {
			go p.tracker.Start()
		}
	}

	p.escalations = &conf.Escalations

	if conf.EventProviders != nil {
		p.providers = *conf.EventProviders
	}

	p.refreshPolicies(conf.Policies)
	p.keepAliveAge = conf.KeepAliveAge
	p.globalPolicy = conf.GlobalPolicy

	if p.globalPolicy != nil {
		p.globalPolicy.Compile()
	}

	// update to the new config
	p.config = conf
	p.Unpause()

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

func (p *Pipeline) Pause() {
	p.pauseChan <- struct{}{}
}

func (p *Pipeline) Unpause() {
	p.unpause()
}

// pause stop processing events
func (p *Pipeline) pause() {
	logrus.Info("Pausing pipeline")

	// make a map to cache the incomming events
	cache := make(map[*event.Event]struct{})
	var e *event.Event
	for {
		select {

		// start caching the events as they come in
		case e = <-p.in:
			logrus.Debugf("Caching event during pause %+v", e)
			cache[e] = struct{}{}

		// when the pause is complete, return control to the caller, and begin sending back
		// the cached events
		case <-p.unpauseChan:

			go func() {

				logrus.Infof("Emptying pause cache of size %d", len(cache))
				// empty the cache
				for e, _ := range cache {
					logrus.Debugf("Proccessing cached event after unpause %+v", *e)
					p.in <- e
				}

				logrus.Info("Pause cache empty complete")

				// signal the unpause function that we are done with the unpause
				p.unpauseChan <- struct{}{}

			}()
			return
		}
	}
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
			p.Pass(e)
			// wait for the event to be done processing before sending the next
			e.Wait()
		}
	}
}

// create keep alive events for each hostname -> time pair
func createKeepAliveEvents(times map[string]time.Time) []*event.Event {
	n := time.Now()
	events := make([]*event.Event, len(times))
	x := 0
	for host, t := range times {
		e := event.NewEvent()
		e.Host = host
		e.Metric = n.Sub(t).Seconds()
		e.Service = KEEP_ALIVE_SERVICE_NAME

		events[x] = e
		x += 1
	}

	return events
}

func (p *Pipeline) Start() {
	logrus.Info("Starting pipeline")

	// fan in all of the providers and process them
	go func() {
		var e *event.Event
		for {
			select {
			// recieve the event
			case e = <-p.in:
				p.Process(e)

			// handle pause
			case <-p.pauseChan:

				// start the pause, and wait until it has been completed
				p.pause()
			}
		}
	}()
}

func (p *Pipeline) ProcessIncident(in *event.Incident) {

	// start tracking this incident in memory so we can call back to it
	p.tracker.TrackIncident(in)
	println(string(in.IndexName()))

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
			logrus.Error("unknown escalation", in.Escalation)
		}
	}
}

// Run the given event though the pipeline
func (p *Pipeline) Process(e *event.Event) {
	if p.globalPolicy != nil {
		if !p.globalPolicy.CheckMatch(e) || !p.globalPolicy.CheckNotMatch(e) {
			return
		}
	}

	// track stas for this event
	p.tracker.TrackEvent(e)

	// process this event on every policy
	var pol *alarm.Policy
	for _, pol = range p.policies {
		e.WaitInc()
		pol.Process(e, func(in *event.Incident) {
			p.ProcessIncident(in)
		})
	}

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
