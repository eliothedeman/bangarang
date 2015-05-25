package api

import (
	"log"
	"net/http"

	"github.com/eliothedeman/bangarang/pipeline"
	"github.com/gorilla/mux"
)

// handles the api methods for incidents
type Incidents struct {
	pipeline *pipeline.Pipeline
}

func NewIncidents(p *pipeline.Pipeline) *Incidents {
	return &Incidents{
		pipeline: p,
	}
}

func (i *Incidents) EndPoint() string {
	return "/api/incidents"
}

// list all the current incidents for the pipeline
func (i *Incidents) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Println(vars)
}
