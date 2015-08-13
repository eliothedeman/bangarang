package api

import (
	"net/http"

	"github.com/eliothedeman/bangarang/pipeline"
	"github.com/gorilla/mux"
)

// handles the api methods for incidents
type Host struct {
	pipeline *pipeline.Pipeline
}

func NewHost(p *pipeline.Pipeline) *Host {
	return &Host{
		pipeline: p,
	}
}

func (h *Host) EndPoint() string {
	return "/api/host/{hostname:.+}"
}

// Delete will resolve a given event
func (h *Host) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	host, ok := vars["hostname"]
	if !ok {
		http.Error(w, "Must append hostname", http.StatusBadRequest)
		return
	}

	h.pipeline.GetTracker().RemoveHost(host)
}

func (h *Host) Get(w http.ResponseWriter, r *http.Request) {
}
