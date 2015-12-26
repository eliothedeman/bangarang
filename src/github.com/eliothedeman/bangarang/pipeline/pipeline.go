package pipeline

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/escalation"
	"github.com/eliothedeman/bangarang/event"
	"github.com/eliothedeman/bangarang/provider"
)

const (
	KEEP_ALIVE_INTERNAL_TAG = "KeepAlive"
)

var (
	DefaultKeepAliveCheckTime = 1 * time.Minute
)

// Pipeline
type Pipeline struct {
	keepAliveAge       time.Duration
	keepAliveCheckTime time.Duration
	escalations        map[string]*escalation.EscalationPolicy
	policies           map[string]*escalation.Policy
	index              *event.Index
	providers          provider.EventProviderCollection
	config             *config.AppConfig
	confLock           sync.Mutex
	tracker            *Tracker
	pauseCache         map[*event.Event]struct{}
	pauseChan          chan struct{}
	unpauseChan        chan struct{}
	in                 chan *event.Event
	incidentInput      chan *event.Incident
}

// NewPipeline returns a pipeline that is empty of any configuation but will still pass events though
func NewPipeline() *Pipeline {
	p := &Pipeline{
		in:                 make(chan *event.Event, 10),
		policies:           make(map[string]*escalation.Policy),
		incidentInput:      make(chan *event.Incident),
		unpauseChan:        make(chan struct{}),
		pauseChan:          make(chan struct{}),
		tracker:            NewTracker(),
		keepAliveCheckTime: DefaultKeepAliveCheckTime,
		escalations:        map[string]*escalation.EscalationPolicy{},
		index:              event.NewIndex(),
	}

	return p
}

func (p *Pipeline) PassEvent(e *event.Event) {
	p.in <- e
}

// only adds polcies that are not already known of
func (p *Pipeline) refreshPolicies(m map[string]*escalation.Policy) {
	logrus.Info("Refreshing policies")

	// initilize the pipeline's polices if they don't already exist
	if p.policies == nil {
		p.policies = make(map[string]*escalation.Policy)
	}

	// add in new policies
	for k, v := range m {

		// compile the new policy
		v.Compile(p)

		// if the name of the new polcy is not known of, insert it
		if _, inMap := p.policies[k]; !inMap {

			logrus.Infof("Adding new policy %s", k)
			p.policies[k] = v
		} else {

			// stop the policy if not. Stops the memory leak
			if p.policies[k] != v {

				// trade the config for the new policies
				m[k] = p.policies[k]

				// stop the old one
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
		logrus.Infof("Stopping policiy %s", name)

		// resolve all incidents that were created by this policy, and are still alive
		ins := p.index.ListIncidents()
		for _, i := range ins {

			// if the incident matched this policy, resolve it
			if i.Policy == name {
				i.Status = event.OK

				// process the resolved incident
				go p.PassIncident(i)
			}
		}
		pol.Stop()
		delete(p.policies, name)
	}
	p.Unpause()
}

// refresh load all config params that don't require a restart
func (p *Pipeline) Refresh(conf *config.AppConfig) {
	p.Pause()
	p.confLock.Lock()
	defer p.confLock.Unlock()

	// if the config has changed at all, refresh the index
	if p.config == nil || string(conf.Hash) != string(p.config.Hash) {
		if !p.tracker.Started() {
			go p.tracker.Start()
		}
	}

	if conf.Escalations != nil {
		p.escalations = conf.Escalations
		// compile each escalation policiy

		for _, v := range p.escalations {
			v.Compile()
		}

	}

	if conf.EventProviders != nil {
		p.providers = *conf.EventProviders
	}

	p.refreshPolicies(conf.Policies)

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

// Pause stops the pipeline and buffers incomming events
func (p *Pipeline) Pause() {
	p.pauseChan <- struct{}{}
}

// Unpause restarts the pipeline and runs the buffered events through the pipeline
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

// GetTracker returns the pipeline's tracker
func (p *Pipeline) GetTracker() *Tracker {
	return p.tracker
}

// checkExpired checks for keep alive
func (p *Pipeline) checkExpired() {
	var events []*event.Event

	// TODO track every tag in the future
	events = createKeepAliveEvents(p.tracker.TagTimes("host"), "host")

	// process every event as if it was an incomming event
	for _, e := range events {
		p.PassEvent(e)
	}
}

// create keep alive events for each hostname -> time pair
func createKeepAliveEvents(times map[string]time.Time, tag string) []*event.Event {
	events := make([]*event.Event, len(times))
	i := 0
	now := time.Now()

	// fill out events
	for k, v := range times {
		e := event.NewEvent()
		e.Tags.Set(tag, k)
		e.Tags.Set(INTERNAL_TAG_NAME, KEEP_ALIVE_INTERNAL_TAG)
		e.Metric = now.Sub(v).Seconds()
		events[i] = e
		i++
	}
	return events
}

// ViewConfig gives access to a read only copy of the current config through a closure
func (p *Pipeline) ViewConfig(f func(c *config.AppConfig)) {

	// lock and get a copy of the current config
	p.confLock.Lock()
	cpy := *p.config
	p.confLock.Unlock()

	// fun the closure on the config copy
	f(&cpy)
}

// UpdateConfig gives access to a copy ofthe pipeline, if no error is returned fron the closure, the pipeline will be refreshed with the given app config
func (p *Pipeline) UpdateConfig(f func(c *config.AppConfig) error, u *config.User) error {

	// lock and make a copy of the config
	p.confLock.Lock()
	cpy := *p.config
	p.confLock.Unlock()

	// run the closure on the config
	err := f(&cpy)

	// return the error if there is one
	if err != nil {
		return err
	}

	// update the config in the db
	cpy.Provider().PutConfig(&cpy, u)

	// if there is no error, refresh the pipeline with the new config
	p.Refresh(&cpy)
	return nil
}

// Start consumes events as they are sent to the pipeline
func (p *Pipeline) Start() {
	logrus.Info("Starting pipeline")

	// fan in all of the providers and process them
	go func() {
		keepAliveCheckTime := time.After(p.keepAliveCheckTime)
		var e *event.Event
		var i *event.Incident
		for {
			select {
			// recieve the event
			case e = <-p.in:
				p.processEvent(e)

			// time to check for keepalives
			case <-keepAliveCheckTime:

				// start the exparation check
				go p.checkExpired()

				// reset the timer for the next check
				keepAliveCheckTime = time.After(p.keepAliveCheckTime)

			case i = <-p.incidentInput:
				p.processIncident(i)

			// handle pause
			case <-p.pauseChan:

				// start the pause, and wait until it has been completed
				p.pause()
			}
		}
	}()
}

// ProcessIncident relays the incident into the pipeline for processing
func (p *Pipeline) PassIncident(in *event.Incident) {
	p.incidentInput <- in
}

// processIncident forwards a deduped incident on to every escalation
func (p *Pipeline) processIncident(in *event.Incident) {
	in.GetEvent().SetState(event.StateIncident)

	// start tracking this incident in memory so we can call back to it
	p.tracker.TrackIncident(in)

	// dedup the incident
	if p.Dedupe(in) {

		// update the incident in the index
		if in.Status != event.OK {
			p.index.PutIncident(in)
		} else {
			p.index.DeleteIncidentById(in.IndexName())
		}

		// send it on to every escalation
		for _, esc := range p.escalations {
			esc.PassIncident(in)
		}
	}

	in.GetEvent().SetState(event.StateComplete)
}

// Run the given event though the pipeline
func (p *Pipeline) processEvent(e *event.Event) {

	// update to current state
	e.SetState(event.StatePipeline)

	// track stas for this event
	p.tracker.TrackEvent(e)

	// process this event on every policy
	var pol *escalation.Policy
	for _, pol = range p.policies {
		pol.PassEvent(e)
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
