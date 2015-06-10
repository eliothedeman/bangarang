package api

import (
	"encoding/json"
	"net/http"

	"github.com/eliothedeman/bangarang/pipeline"
)

// handles the api methods for incidents
type KnownHosts struct {
	pipeline *pipeline.Pipeline
}

func NewKnownHosts(pipe *pipeline.Pipeline) *KnownHosts {
	return &KnownHosts{
		pipeline: pipe,
	}
}

func (e *KnownHosts) EndPoint() string {
	return "/api/stats/hosts"
}

func (e *KnownHosts) Get(w http.ResponseWriter, r *http.Request) {
	t := e.pipeline.GetTracker()

	buff, err := json.Marshal(t.GetHosts())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(buff)
}
