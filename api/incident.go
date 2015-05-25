package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/eliothedeman/bangarang/pipeline"
	"github.com/gorilla/mux"
	"github.com/pquerna/ffjson/ffjson"
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
	return "/api/incident/{id:[0-9]+}"
}

func (i *Incident) Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	log.Println("[]")
}

// list all the current incidents for the pipeline
func (i *Incident) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	vars := mux.Vars(r)

	if id, ok := vars["id"]; !ok {
		log.Println(MUST_INCLUDE_ID.Error())
		http.Error(w, MUST_INCLUDE_ID.Error(), http.StatusBadRequest)
		return
	} else {
		// all is well
		incidentId, _ := strconv.Atoi(id)
		incident := i.pipeline.GetIncident(int64(incidentId))
		if incident == nil {
			w.Write([]byte("{}"))
			return
		}

		// encode incident
		buff, err := ffjson.MarshalFast(incident)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}

		w.Write(buff)
	}
}
