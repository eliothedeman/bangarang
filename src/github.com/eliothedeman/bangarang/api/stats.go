package api

import (
	"encoding/json"
	"net/http"

	"github.com/eliothedeman/bangarang/pipeline"
)

// handles the api methods for incidents
type EventStats struct {
	pipeline *pipeline.Pipeline
}

func NewEventStats(pipe *pipeline.Pipeline) *EventStats {
	return &EventStats{
		pipeline: pipe,
	}
}

func (e *EventStats) EndPoint() string {
	return "/api/stats/event"
}

func (e *EventStats) Get(w http.ResponseWriter, r *http.Request) {
	t := e.pipeline.GetTracker()
	report := t.GetStats()

	buff, err := json.Marshal(report)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(buff)
}
