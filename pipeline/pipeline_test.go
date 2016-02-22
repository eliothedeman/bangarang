package pipeline

import (
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/escalation"
	"github.com/eliothedeman/bangarang/escalation/test"
	"github.com/eliothedeman/bangarang/event"
)

var (
	tests_ran = 100
)

type testContext struct {
	before, after func(p *Pipeline)
	pipeline      *Pipeline
}

func baseTestContext() *testContext {
	c := &testContext{
		pipeline: NewPipeline(),
		before: func(p *Pipeline) {

		},
		after: func(p *Pipeline) {
			p.index.Delete()
		},
	}

	return c
}

func (t *testContext) addPolicy(name string, pol *escalation.Policy) {

}

func runningTestContext() *testContext {

	b := baseTestContext()
	b.before = func(p *Pipeline) {
		p.Start()
		conf := config.NewDefaultConfig()
		conf.SetProvider(config.NewMockProvider())
		p.Refresh(conf)
	}

	return b
}

func (t *testContext) runTest(f func(p *Pipeline)) {
	t.before(t.pipeline)
	defer t.after(t.pipeline)
	f(t.pipeline)
}

func (t *testContext) getCurrentConfig() *config.AppConfig {
	var c *config.AppConfig
	t.pipeline.ViewConfig(func(conf *config.AppConfig) {
		c = conf
	})

	return c
}
func (t *testContext) addEscalationPolicy(name string, p *escalation.EscalationPolicy) {
	c := t.getCurrentConfig()
	if c.Escalations == nil {
		c.Escalations = make(map[string]*escalation.EscalationPolicy)
	}
	c.Escalations[name] = p

	// refresh
	t.pipeline.Refresh(c)
}

func (t *testContext) start() {
	logrus.SetLevel(logrus.DebugLevel)
	t.pipeline.Start()
}

// end the current test
func (t *testContext) end() {
	t.pipeline.index.Delete()
}

type testingPasser struct {
	incidents map[string]*event.Incident
}

func (t *testingPasser) PassIncident(i *event.Incident) {
	if t.incidents == nil {
		t.incidents = map[string]*event.Incident{}
	}
	t.incidents[string(i.IndexName())] = i
}

func newTestPasser() event.IncidentPasser {
	return &testingPasser{}
}

func testCondition(g, l, e *float64, o int) *escalation.Condition {
	return &escalation.Condition{
		Greater:    g,
		Less:       l,
		Exactly:    e,
		Occurences: o,
	}
}

func test_f(f float64) *float64 {
	return &f
}

func TestNewPipeline(t *testing.T) {
	x := baseTestContext()
	x.runTest(func(n *Pipeline) {
		if n.escalations == nil {
			t.Fatal()
		}
		if n.policies == nil {
			t.Fatal()
		}

		if n.index == nil {
			t.Fatal()
		}

		if n.config != nil {
			t.Fatal()
		}

		if n.pauseChan == nil {
			t.Fatal()
		}
		if n.unpauseChan == nil {
			t.Fatal()
		}
		if n.in == nil {
			t.Fatal()
		}
		if n.incidentInput == nil {
			t.Fatal()
		}
	})
}

func TestPassEvent(t *testing.T) {
	x := baseTestContext()
	x.runTest(func(p *Pipeline) {
		e := event.NewEvent()
		go func() {
			p.PassEvent(e)
		}()

		ne := <-p.in
		if ne != e {
			t.Fatal()
		}
	})
}

func TestPauseUnpause(t *testing.T) {
	x := runningTestContext()
	x.runTest(func(p *Pipeline) {
		// make sure events can go in and out
		e := event.NewEvent()
		p.PassEvent(e)
		p.Pause()

		for i := 0; i < 1000; i++ {
			p.PassEvent(event.NewEvent())
		}

		p.Unpause()

		// pass another event
		p.PassEvent(event.NewEvent())
	})
}

func TestRemovePolicy(t *testing.T) {
	x := runningTestContext()
	x.runTest(func(p *Pipeline) {
		// make a test user
		u := config.NewUser("test", "test", "password", config.WRITE)

		// add an empty escalation policy
		p.UpdateConfig(func(c *config.AppConfig) error {
			c.Policies["hello"] = &escalation.Policy{}
			c.Policies["hello"].Compile(&testingPasser{})

			return nil
		}, u)

		// make sure the policy is there
		p.ViewConfig(func(c *config.AppConfig) {
			if _, ok := c.Policies["hello"]; !ok {
				t.Fatal("Policy was not added")
			}
		})

		// remove policy
		p.UpdateConfig(func(c *config.AppConfig) error {
			delete(c.Policies, "hello")
			return nil
		}, u)

		// make sure the policy is gone
		p.ViewConfig(func(c *config.AppConfig) {
			if _, ok := c.Policies["hello"]; ok {
				t.Fatal("Policy was not removed")
			}
		})
	})
}

func TestGetTracker(t *testing.T) {
	x := baseTestContext()
	x.runTest(func(p *Pipeline) {
		track := p.GetTracker()
		if track == nil {
			t.Fatal()
		}
	})
}

func TestRefresh(t *testing.T) {
	x := runningTestContext()
	x.runTest(func(p *Pipeline) {
		p.PassEvent(event.NewEvent())
		p.Refresh(config.NewDefaultConfig())
		p.PassEvent(event.NewEvent())
	})
}

func TestPutIncident(t *testing.T) {
	x := runningTestContext()
	x.runTest(func(p *Pipeline) {
		in := event.NewIncident("test", event.OK, event.NewEvent())
		p.PutIncident(in)
	})
}

func TestListIncidents(t *testing.T) {
	x := runningTestContext()
	x.runTest(func(p *Pipeline) {
		in := event.NewIncident("test", event.OK, event.NewEvent())
		p.PutIncident(in)

		if l := p.ListIncidents(); l[0].Policy != in.Policy {
			t.Fatal("Incident was not added to the index")
		}
	})
}

func TestDedupe(t *testing.T) {
	x := runningTestContext()
	x.runTest(func(p *Pipeline) {
		in := event.NewIncident("test", event.OK, event.NewEvent())
		p.PutIncident(in)

		if p.Dedupe(in) {
			t.Fatal("Incident was not deduped")
		}
	})
}

func TestKeepAlive(t *testing.T) {
	x := runningTestContext()
	x.runTest(func(p *Pipeline) {

		u := &config.User{}
		u.Permissions = config.WRITE

		ta := test.NewTestAlert()

		// add a escalation policy that will catch a keep alive
		p.UpdateConfig(func(c *config.AppConfig) error {
			esc := &escalation.EscalationPolicy{}
			esc.Crit = true
			esc.Escalations = []escalation.Escalation{ta}
			esc.Match = event.NewTagset(0)
			esc.Match.Set(INTERNAL_TAG_NAME, ".*")
			c.Escalations["test"] = esc
			esc.Compile()

			// create a condition for the KeepAlive
			pol := &escalation.Policy{}
			pol.Match = event.NewTagset(0)
			pol.Match.Set(INTERNAL_TAG_NAME, KEEP_ALIVE_INTERNAL_TAG)

			cond := &escalation.Condition{}
			cond.Simple = true
			cond.Greater = test_f(0)
			cond.WindowSize = 2
			cond.Occurences = 1
			pol.Crit = cond
			c.Policies["test"] = pol

			return nil

		}, u)

		// create an event
		e := event.NewEvent()
		e.Tags.Set("host", "test")
		p.PassEvent(e)
		e.WaitForState(event.StateComplete, time.Millisecond)

		time.Sleep(20 * time.Millisecond)
		p.checkExpired()
		time.Sleep(50 * time.Millisecond)

		if len(ta.Incidents) != 1 {
			t.Error(ta.Incidents)
		}

		ka := ta.Incidents[0].GetEvent()
		if ka.Get("host") != "test" {
			t.Fail()
		}

		if ka.Get(INTERNAL_TAG_NAME) != KEEP_ALIVE_INTERNAL_TAG {
			t.Fail()
		}
	})
}

func TestResolve(t *testing.T) {
	x := runningTestContext()
	x.runTest(func(p *Pipeline) {
		u := &config.User{}
		u.Permissions = config.WRITE

		ta := test.NewTestAlert()

		// add a escalation policy that will catch our metrics
		p.UpdateConfig(func(c *config.AppConfig) error {
			esc := &escalation.EscalationPolicy{}
			esc.Crit = true
			esc.Ok = true
			esc.Escalations = []escalation.Escalation{ta}
			esc.Match = event.NewTagset(0)
			esc.Match.Set("host", ".*")
			c.Escalations["test"] = esc
			esc.Compile()

			// create a policy that will match against this host
			pol := &escalation.Policy{}
			pol.Match = event.NewTagset(0)
			pol.Match.Set("host", ".*")

			cond := &escalation.Condition{}
			cond.Simple = true
			cond.Greater = test_f(1)
			cond.WindowSize = 2
			cond.Occurences = 1
			pol.Crit = cond
			c.Policies["test"] = pol

			return nil

		}, u)

		// create an event
		e := event.NewEvent()

		e.Metric = 4
		e.Tags.Set("host", "test")
		p.PassEvent(e)
		e.WaitForState(event.StateComplete, time.Millisecond)

		// TODO get rid of waiting for things to pass through the pipeline
		time.Sleep(50 * time.Millisecond)
		if len(ta.Incidents) != 1 {
			t.Error(ta.Incidents)
		}

		time.Sleep(50 * time.Millisecond)

		e.Metric = 0
		p.PassEvent(e)
		e.WaitForState(event.StateComplete, time.Millisecond)
		time.Sleep(50 * time.Millisecond)
		if len(ta.Incidents) != 2 {
			t.Error(ta.Incidents[0])
		}
	})
}
