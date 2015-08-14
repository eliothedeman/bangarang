package api

import (
	"net/http"

	"github.com/eliothedeman/bangarang/pipeline"
)

// handles the api methods for incidents
type IncidentGraph struct {
	pipeline *pipeline.Pipeline
}

func NewIncidentGraph(p *pipeline.Pipeline) *IncidentGraph {
	return &IncidentGraph{
		pipeline: p,
	}
}

func (i *IncidentGraph) EndPoint() string {
	return "/api/incident/graph/{id:.+}"
}

func (i *IncidentGraph) Get(w http.ResponseWriter, r *http.Request) {

	// p, _ := plot.New()
	// p.Y.Label.Text = fmt.Sprintf("%s %s", e.Service, e.SubService)
	// p.X.Label.Text = "Time"
	// plotutil.AddLinePoints(p, context.PlotPoints())
	// w, _ := p.WriterTo(16*vg.Inch, 9*vg.Inch, png)
}
