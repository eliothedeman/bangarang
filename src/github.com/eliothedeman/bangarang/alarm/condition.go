package alarm

import (
	"math"
	"regexp"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/event"
	"github.com/eliothedeman/smoothie"
)

var (
	DEFAULT_WINDOW_SIZE = 100 // The default size of the dataframe used in window operations
	STATUS_SIZE         = 10  // The default size of the dataframe used to count statuses
)

// Condition holds conditional information to check events against
type Condition struct {
	Greater       *float64     `json:"greater"`
	Less          *float64     `json:"less"`
	Exactly       *float64     `json:"exactly"`
	StdDev        *StdDev      `json:"std_dev"`
	Escalation    string       `json:"escalation"`
	Occurences    int          `json:"occurences"`
	WindowSize    int          `json:"window_size"`
	Aggregation   *Aggregation `json:"agregation"`
	trackFunc     TrackFunc
	groupBy       grouper
	checks        []satisfier
	eventTrackers map[string]*eventTracker
	sync.Mutex
	ready bool
}

// Config for checks based on the aggrigation of data over a time window, instead of individual data points
type Aggregation struct {
	WindowLength int `json:"window_length"`
}

// Config for checks based on standard deviation
type StdDev struct {
	Sigma      float64 `json:"sigma"`
	WindowSize *int    `json:"window_size"`
}

type aggregator struct {
	nextCloseout time.Time
}

type matcher struct {
	name  string
	match *regexp.Regexp
}

type grouper []*matcher

// generate an index name by using group-by statements
func (g grouper) genIndexName(e *event.Event) string {
	name := ""
	for _, m := range g {
		res := m.match.FindStringSubmatch(e.Get(m.name))

		switch len(res) {
		case 1:
			name = name + ":" + res[0]
		case 2:
			name = name + ":" + res[1]
			//
		}
	}
	return name
}

type eventTracker struct {
	df         *smoothie.DataFrame
	states     *smoothie.DataFrame
	occurences int

	// optional
	agg *aggregator
}

type satisfier func(e *event.Event) bool

func (c *Condition) newTracker() *eventTracker {
	et := &eventTracker{
		df:     smoothie.NewDataFrame(c.WindowSize),
		states: smoothie.NewDataFrameFromSlice(make([]float64, STATUS_SIZE)),
	}

	if c.Aggregation != nil {
		et.agg = &aggregator{}
	}

	return et
}

func (c *Condition) DoOnTracker(e *event.Event, dot func(*eventTracker)) {
	c.Lock()
	et, ok := c.eventTrackers[c.groupBy.genIndexName(e)]
	if !ok {
		et = c.newTracker()
		c.eventTrackers[c.groupBy.genIndexName(e)] = et
	}
	dot(et)
	c.Unlock()
}

type TrackFunc func(c *Condition, e *event.Event) bool

func AggregationTrack(c *Condition, e *event.Event) bool {
	c.DoOnTracker(e, func(t *eventTracker) {

		// if we are still within the closeout, add to the current value
		if time.Now().Before(t.agg.nextCloseout) {
			t.df.Insert(0, t.df.Index(0)+e.Metric)

			// if we are after the closeout, start a new datapoint and close out the old one
		} else {
			t.df.Push(e.Metric)
			t.agg.nextCloseout = time.Now().Add(time.Second * time.Duration(c.Aggregation.WindowLength))
		}
	})

	return c.OccurencesHit(e)
}

func SimpleTrack(c *Condition, e *event.Event) bool {
	c.DoOnTracker(e, func(t *eventTracker) {
		t.df.Push(e.Metric)
	})

	return c.OccurencesHit(e)
}

// start tracking an event, and returns if the event has hit it's occurence settings
func (c *Condition) TrackEvent(e *event.Event) bool {
	return c.trackFunc(c, e)
}

func (c *Condition) StateChanged(e *event.Event) bool {
	changed := false
	c.DoOnTracker(e, func(t *eventTracker) {
		changed = t.states.Index(0) == t.states.Index(1)
	})
	return changed
}

// check to see if an event has it the occurences level
func (c *Condition) OccurencesHit(e *event.Event) bool {
	occ := 0

	if c.Satisfies(e) {
		c.DoOnTracker(e, func(t *eventTracker) {
			t.occurences += 1
			occ = t.occurences
			t.states.Push(1)
		})
	} else {
		c.DoOnTracker(e, func(t *eventTracker) {
			t.occurences = 0
			t.states.Push(0)
		})
	}

	return occ >= c.Occurences
}

// check if an event satisfies a condition
func (c *Condition) Satisfies(e *event.Event) bool {
	for _, check := range c.checks {
		if check(e) {
			return true
		}
	}

	return false
}

// create a list of checks that the condition will use to test events
func (c *Condition) compileChecks() []satisfier {
	s := []satisfier{}

	if c.Greater != nil {
		logrus.Info("Adding greater than check:", *c.Greater)
		s = append(s, func(e *event.Event) bool {
			return e.Metric > *c.Greater
		})
	}
	if c.Less != nil {
		logrus.Info("Adding less than check:", *c.Less)
		s = append(s, func(e *event.Event) bool {
			return e.Metric < *c.Less
		})
	}
	if c.Exactly != nil {
		logrus.Info("Adding exactly check:", *c.Exactly)
		s = append(s, func(e *event.Event) bool {
			return e.Metric == *c.Exactly
		})
	}

	if c.StdDev != nil {
		logrus.Infof("Adding a standard deviation check: %+v", *c.StdDev)
		s = append(s, func(e *event.Event) bool {
			met := false
			c.DoOnTracker(e, func(t *eventTracker) {
				avg := t.df.Avg()
				if math.Abs(e.Metric-avg) > t.df.StdDev()*c.StdDev.Sigma {
					met = true
				}
			})
			return met
		})
	}

	// if we are using aggregation, replace all with the aggregation form
	if c.Aggregation != nil {
		logrus.Info("Converting checks to using aggregation")
		for i := range s {
			s[i] = c.wrapAggregation(s[i])
		}
	}
	return s
}

func (c *Condition) wrapAggregation(s satisfier) satisfier {
	return func(e *event.Event) bool {
		// create a new event with the aggregated value
		ne := *e
		c.DoOnTracker(e, func(t *eventTracker) {
			ne.Metric = t.df.Index(0)
		})

		return s(&ne)
	}
}

func compileGrouper(gb map[string]string) grouper {
	g := grouper{}
	for k, v := range gb {
		g = append(g, &matcher{name: k, match: regexp.MustCompile(v)})
	}
	return g
}

func getTrackingFunc(c *Condition) TrackFunc {
	if c.Aggregation != nil {
		return AggregationTrack
	}

	return SimpleTrack
}

func (c *Condition) init(groupBy map[string]string) {
	c.groupBy = compileGrouper(groupBy)

	c.checks = c.compileChecks()

	// fixes issue where occurences are hit, even when the event doesn't satisify the condition
	if c.Occurences < 1 {
		logrus.Warnf("Occurences must be > 1. %d given. Occurences for this condition will be set to 1.", c.Occurences)
		c.Occurences = 1
	}

	if c.eventTrackers == nil {
		c.eventTrackers = make(map[string]*eventTracker)
	}

	if c.WindowSize == 0 {
		c.WindowSize = DEFAULT_WINDOW_SIZE
	}

	// decide which tracking method we will use
	c.trackFunc = getTrackingFunc(c)

	c.ready = true
}
