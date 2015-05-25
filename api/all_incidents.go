package api

import (
	"log"
	"net/http"

	"github.com/eliothedeman/bangarang/pipeline"
	"github.com/pquerna/ffjson/ffjson"
)

// handles the api methods for incidents
type AllIncidents struct {
	pipeline *pipeline.Pipeline
}

func NewAllIncidents(p *pipeline.Pipeline) *AllIncidents {
	return &AllIncidents{
		pipeline: p,
	}
}

func (a *AllIncidents) EndPoint() string {
	return "/api/all-incidents"
}

// list all the current incidents for the pipeline
func (a *AllIncidents) Get(w http.ResponseWriter, r *http.Request) {
	ins := a.pipeline.ListIncidents()

	buff, err := ffjson.Marshal(ins)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(buff)
}
