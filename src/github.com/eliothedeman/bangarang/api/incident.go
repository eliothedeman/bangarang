package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/event"
	"github.com/eliothedeman/bangarang/pipeline"
	"github.com/gorilla/mux"
)

var (
	MUST_INCLUDE_ID = errors.New("Must include id")
)

// handles the api methods for incidents
type Incident struct {
	pipeline *pipeline.Pipeline
}

func NewIncident(p *pipeline.Pipeline) *Incident {
	return &Incident{
		pipeline: p,
	}
}

func (i *Incident) EndPoint() string {
	return "/api/incident/{id}"
}

// return any incidnet that is greater than this value
func reduceStatusAbove(level int, in []*event.Incident) []*event.Incident {
	out := []*event.Incident{}

	if in == nil {
		return out
	}
	for _, i := range in {
		if i.Status >= level {
			out = append(out, i)
		}
	}

	return out
}

func makeKV(in []*event.Incident) map[string]*event.Incident {
	out := map[string]*event.Incident{}
	for _, i := range in {
		out[string(i.IndexName())] = i
	}
	return out
}

// Create an incident
func (i *Incident) Post(req *Request) {
	buff, err := ioutil.ReadAll(req.r.Body)
	if err != nil {
		http.Error(req.w, err.Error(), http.StatusBadRequest)
		return
	}

	in := &event.Incident{}
	err = json.Unmarshal(buff, in)
	if err != nil {
		http.Error(req.w, err.Error(), http.StatusBadRequest)
		return
	}

	i.pipeline.PassIncident(in)
}

// Delete will resolve a given event
func (i *Incident) Delete(req *Request) {
	vars := mux.Vars(req.r)
	id, ok := vars["id"]
	if !ok {
		http.Error(req.w, "Must append incident id", http.StatusBadRequest)
		return
	}

	index := i.pipeline.GetIndex()
	in := index.GetIncident([]byte(id))

	if in == nil {
		logrus.Errorf("Incident %s not found", id)
		http.Error(req.w, fmt.Sprintf("Incident %s not found", id), http.StatusInternalServerError)
		return
	}

	// fetch the callback channel to resolve this incident
	res := i.pipeline.GetTracker().GetIncidentResolver(in)

	// if we have a non-nil channel, this incident was created during this run
	// which means we can clear it's state
	if res != nil {

		// make a copy
		x := *in

		// send the incident on so it can be used to clear state in the policy
		// which created it
		res <- &x
	}

	// if an incident with this id exists, set it's status to ok and send it back through the pipeline
	if in != nil {
		in.Status = event.OK
		in.Description = ""
		i.pipeline.PassIncident(in)
	}
}

func (i *Incident) Get(req *Request) {
	req.w.Header().Add("content-type", "application/json")
	vars := mux.Vars(req.r)
	id, ok := vars["id"]
	if !ok {
		http.Error(req.w, "Must append incident id", http.StatusBadRequest)
		return
	}
	index := i.pipeline.GetIndex()

	// if the id is "*", fetch all outstanding incidents
	if id == "*" {
		all := index.ListIncidents()
		all = reduceStatusAbove(event.WARNING, all)
		buff, err := json.Marshal(makeKV(all))
		if err != nil {
			logrus.Error(err)
			http.Error(req.w, err.Error(), http.StatusInternalServerError)
			return
		}
		req.w.Write(buff)
		return
	}

	// write out the found incident. The value will be null if nothing was found
	in := index.GetIncident([]byte(id))
	buff, err := json.Marshal(in)
	if err != nil {
		logrus.Error(err)
		http.Error(req.w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.w.Write(buff)
}
