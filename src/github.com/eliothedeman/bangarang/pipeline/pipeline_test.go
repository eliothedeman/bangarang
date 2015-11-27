package pipeline

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/escalation"
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
	return &testContext{
		pipeline: NewPipeline(),
		before: func(p *Pipeline) {

		},
		after: func(p *Pipeline) {
			p.index.Delete()
		},
	}
}

func (t *testContex) addPolicy(name string, pol *escalation.Policy) {

}

func runningTestContext() *testContext {
	return &testContext{
		pipeline: NewPipeline(),
		before: func(p *Pipeline) {
			p.Start()

		},
		after: func(p *Pipeline) {
			p.index.Delete()
		},
	}
}

func (t *testContext) runTest(f func(p *Pipeline)) {
	t.before(t.pipeline)
	f(t.pipeline)
	t.after(t.pipeline)
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
}
