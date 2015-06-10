package api

import (
	"encoding/json"
	"net/http"

	"github.com/eliothedeman/bangarang/pipeline"
)

// handles the api methods for incidents
type KnownServices struct {
	pipeline *pipeline.Pipeline
}

func NewKnownServices(pipe *pipeline.Pipeline) *KnownServices {
	return &KnownServices{
		pipeline: pipe,
	}
}

func (e *KnownServices) EndPoint() string {
	return "/api/stats/services"
}

func (e *KnownServices) Get(w http.ResponseWriter, r *http.Request) {
	t := e.pipeline.GetTracker()

	buff, err := json.Marshal(t.GetServices())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(buff)
}
