package escalation

import (
	"math"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/event"
	"github.com/eliothedeman/smoothie"
)

const (
	DEFAULT_WINDOW_SIZE     = 2  // The default size of the dataframe used in window operations
	STATUS_SIZE             = 10 // The default size of the dataframe used to count statuses
	MIN_STD_DEV_WINDOW_SIZE = 5  // the smallets a window size can be for a standard deviation check
)

// Condition holds conditional information to check events against
type Condition struct {
	Greater       *float64     `json:"greater"`
	Less          *float64     `json:"less"`
	Exactly       *float64     `json:"exactly"`
	StdDev        bool         `json:"std_dev"`
	Derivative    bool         `json:"derivative"`
	HoltWinters   bool         `json:"holt_winters"`
	Simple        bool         `json:"simple"`
	Occurences    int          `json:"occurences"`
	WindowSize    int          `json:"window_size"`
	Aggregation   *Aggregation `json:"agregation"`
	trackFunc     TrackFunc
	checks        []satisfier
	eventTrackers map[string]*eventTracker
	sync.Mutex
	ready bool
}

// Config for checks based on the aggrigation of data over a time window, instead of individual data points
type Aggregation struct {
	WindowLength int `json:"window_length"`
}

type aggregator struct {
	nextCloseout time.Time
}

type eventTracker struct {
	df         *smoothie.DataFrame
	states     *smoothie.DataFrame
	count      int
	occurences int

	// optional
	agg *aggregator
}

func (e *eventTracker) refresh() {
	e.states = smoothie.NewDataFrameFromSlice(make([]float64, STATUS_SIZE))
	e.occurences = 0
}

type satisfier func(e *event.Event) bool

func (c *Condition) newTracker() *eventTracker {
	et := &eventTracker{
		df:     smoothie.NewDataFrameFromSlice(make([]float64, c.WindowSize)),
		states: smoothie.NewDataFrameFromSlice(make([]float64, STATUS_SIZE)),
	}

	if c.Aggregation != nil {
		et.agg = &aggregator{}
	}

	return et
}

func (c *Condition) DoOnTracker(e *event.Event, dot func(*eventTracker)) {
	// c.Lock()
	et, ok := c.eventTrackers[e.IndexName()]
	if !ok {
		et = c.newTracker()
		c.eventTrackers[e.IndexName()] = et
	}
	dot(et)
	// c.Unlock()
}

func (c *Condition) getTracker(e *event.Event) *eventTracker {
	if c.eventTrackers == nil {
		c.eventTrackers = make(map[string]*eventTracker)
	}
	et, ok := c.eventTrackers[e.IndexName()]
	if !ok {
		et = c.newTracker()
		c.eventTrackers[e.IndexName()] = et
	}

	return et
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
	t := c.getTracker(e)
	t.df.Push(e.Metric)
	t.count += 1

	return c.OccurencesHit(e)
}

// check to see if this condition should be treated as simple
func (c *Condition) isSimple() bool {
	if c.Simple {
		return true
	}

	// if nothing is set, default to simple
	if !(c.StdDev || c.HoltWinters || c.Derivative) {
		return true
	}
	return false
}

// start tracking an event, and returns if the event has hit it's occurence settings
func (c *Condition) TrackEvent(e *event.Event) bool {
	return c.trackFunc(c, e)
}

// StateChanged returns true if the state of the incoming event is not the same as the last event
func (c *Condition) StateChanged(e *event.Event) bool {
	t := c.getTracker(e)
	if t.count == 0 && t.states.Index(t.states.Len()-1) != 0 {
		return true
	}
	return t.states.Index(t.states.Len()-1) != t.states.Index(t.states.Len()-2)
}

// check to see if an event has hit the occurences level
func (c *Condition) OccurencesHit(e *event.Event) bool {

	t := c.getTracker(e)

	if c.Satisfies(e) {
		t.occurences += 1
	} else {
		t.occurences = 0
	}

	if t.occurences >= c.Occurences {
		t.states.Push(1)
	} else {
		t.states.Push(0)
	}

	return t.occurences >= c.Occurences
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

	// if any of the special checks are included, only one check can be implemented per condition
	if !c.isSimple() {
		if c.StdDev {

			// check to see if the window size is large enough. The minimum 5.
			if c.WindowSize < MIN_STD_DEV_WINDOW_SIZE {
				logrus.Errorf("Window size %d is too small for a standard deviation check", c.WindowSize)

				// stop short
				return s
			}

			sigma := math.NaN()
			// get the sigma value
			if c.Greater != nil {
				sigma = *c.Greater
			} else {
				// default to 5 sigma
				sigma = 5
			}

			logrus.Infof("Adding standard deviation check of %f sigma", sigma)

			s = append(s, func(e *event.Event) bool {
				t := c.getTracker(e)

				// if the count is greater than 1/4 the window size, start checking
				if t.count > t.df.Len()/4 {

					// if the count is greater than the window size, use the whole df
					if t.count >= t.df.Len() {
						return math.Abs(e.Metric-t.df.Avg()) > (sigma * t.df.StdDev())
					}

					// take a sublslice of populated values
					return math.Abs(e.Metric-t.df.Index(t.df.Len()-2)) > (sigma * t.df.StdDev())
				}
				return false
			})
			return s
		}

		if c.Derivative {
			check := math.NaN()
			var kind uint8
			// get the check value
			if c.Greater != nil {
				kind = 1
				check = *c.Greater
			} else if c.Less != nil {
				kind = 2
				check = *c.Less
			} else if c.Exactly != nil {
				kind = 3
				check = *c.Exactly
			} else {
				logrus.Error("No derivitive type supplied. >, <, == required")
			}

			if kind != 0 {
				logrus.Infof("Adding derivative check of %f", check)
				s = append(s, func(e *event.Event) bool {
					t := c.getTracker(e)

					// we need to have seen at least enough events to
					if t.count < t.df.Len() {
						return false
					}

					diff := e.Metric - t.df.Index(0)
					switch kind {
					case 1:
						return diff > check

					case 2:
						return diff < check

					case 3:
						return diff == check
					}
					return false
				})
			}

			return s
		}

	} else {
		if c.Greater != nil {
			logrus.Info("Adding greater than check:", *c.Greater)
			gt := *c.Greater
			s = append(s, func(e *event.Event) bool {
				return e.Metric > gt
			})
		}
		if c.Less != nil {
			logrus.Info("Adding less than check:", *c.Less)
			lt := *c.Less
			s = append(s, func(e *event.Event) bool {
				return e.Metric < lt
			})
		}
		if c.Exactly != nil {
			logrus.Info("Adding exactly check:", *c.Exactly)
			ex := *c.Exactly
			s = append(s, func(e *event.Event) bool {
				return e.Metric == ex
			})
		}
	}

	// if we are using aggregation, replace all with the aggregation form
	if c.Aggregation != nil {
		logrus.Infof("Converting %d checks to using aggregation", len(s))
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

func getTrackingFunc(c *Condition) TrackFunc {
	if c.Aggregation != nil {
		return AggregationTrack
	}

	return SimpleTrack
}

// init compiles checks and sanatizes the conditon before returning itself
func (c *Condition) init(groupBy *event.TagSet) {
	c.checks = c.compileChecks()
	c.eventTrackers = make(map[string]*eventTracker)

	// fixes issue where occurences are hit, even when the event doesn't satisify the condition
	if c.Occurences < 1 {
		logrus.Warnf("Occurences must be > 1. %d given. Occurences for this condition will be set to 1.", c.Occurences)
		c.Occurences = 1
	}

	// if we have no trackers already, make an empty map of them
	if c.eventTrackers == nil {
		c.eventTrackers = make(map[string]*eventTracker)
	}

	// WindowSize must be above 2. At least one piece of data is needed for historical checks.
	if c.WindowSize < 2 {
		logrus.Warnf("WindowSize must be >= 1. %d given. Window size for this condition will be set to %d", c.WindowSize, DEFAULT_WINDOW_SIZE)
		c.WindowSize = DEFAULT_WINDOW_SIZE
	}

	// decide which tracking method we will use
	c.trackFunc = getTrackingFunc(c)

	c.ready = true
}
