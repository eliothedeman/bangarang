package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/escalation"
	"github.com/eliothedeman/bangarang/pipeline"
	"github.com/gorilla/mux"
)

// EscalationConfig handles the api methods for incidents
type EscalationConfig struct {
	pipeline *pipeline.Pipeline
}

func NewEscalationConfig(pipe *pipeline.Pipeline) *EscalationConfig {
	return &EscalationConfig{
		pipeline: pipe,
	}
}

// EndPoint return the endpoint of this method
func (p *EscalationConfig) EndPoint() string {
	return "/api/escalation/config/{id}"
}

// Get HTTP get method
func (p *EscalationConfig) Get(req *Request) {
	var conf *config.AppConfig
	p.pipeline.ViewConfig(func(cf *config.AppConfig) {
		conf = cf
	})
	vars := mux.Vars(req.r)
	id, ok := vars["id"]
	if !ok {
		logrus.Error("Must append escalation id", req.r.URL.String())
		http.Error(req.w, "must append escalation id", http.StatusBadRequest)
		return
	}

	if id == "*" {
		buff, err := json.Marshal(&conf.Escalations)
		if err != nil {
			logrus.Error(err)
			http.Error(req.w, err.Error(), http.StatusBadRequest)
			return
		}
		req.w.Write(buff)
		return
	}

}

// Delete the given event provider
func (p *EscalationConfig) Delete(req *Request) {

	err := p.pipeline.UpdateConfig(func(conf *config.AppConfig) error {
		vars := mux.Vars(req.r)
		id, ok := vars["id"]
		if !ok {
			return fmt.Errorf("Must append escalation id %s", req.r.URL)
		}

		logrus.Info("Removing escalation: %s", id)
		delete(conf.Escalations, id)
		return nil

	}, req.u)

	if err != nil {
		logrus.Error(err)
		http.Error(req.w, err.Error(), http.StatusBadRequest)
	}
}

// Post HTTP get method
func (p *EscalationConfig) Post(req *Request) {

	err := p.pipeline.UpdateConfig(func(conf *config.AppConfig) error {
		vars := mux.Vars(req.r)

		id, ok := vars["id"]
		if !ok {
			return fmt.Errorf("Must append escalation id %s", req.r.URL)
		}

		buff, err := ioutil.ReadAll(req.r.Body)
		if err != nil {
			return err
		}

		esc := &escalation.EscalationPolicy{}
		err = json.Unmarshal(buff, esc)
		if err != nil {
			return err
		}

		conf.Escalations[id] = esc

		return nil

	}, req.u)

	if err != nil {
		logrus.Error(err)
		http.Error(req.w, err.Error(), http.StatusBadRequest)

	}

}
