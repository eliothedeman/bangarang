package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/pipeline"
	"github.com/gorilla/mux"
)

const (
	global_name = "global"
)

// PolicyConfig handles the api methods for incidents
type PolicyConfig struct {
	pipeline *pipeline.Pipeline
}

func NewPolicyConfig(pipe *pipeline.Pipeline) *PolicyConfig {
	return &PolicyConfig{
		pipeline: pipe,
	}
}

// EndPoint return the endpoint of this method
func (p *PolicyConfig) EndPoint() string {
	return "/api/policy/config/{id}"
}

// Get HTTP get method
func (p *PolicyConfig) Get(w http.ResponseWriter, r *http.Request) {
	conf := p.pipeline.GetConfig()
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		logrus.Error("Must append policy id", r.URL.String())
		http.Error(w, "must append policy id", http.StatusBadRequest)
		return
	}

	if id == "*" {
		buff, err := json.Marshal(conf.Policies)
		if err != nil {
			logrus.Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Write(buff)
		return
	}

	// special case for global
	var pol *alarm.Policy
	if id == "global" {
		pol = conf.GlobalPolicy
	} else {
		pol, ok = conf.Policies[id]
		if !ok {
			http.Error(w, fmt.Sprintf("Unable to find policy '%s'", id), http.StatusBadRequest)
			return
		}
	}
	buff, err := json.Marshal(pol)
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write(buff)
}

// Delete the given event provider
func (p *PolicyConfig) Delete(w http.ResponseWriter, r *http.Request) {
	conf := p.pipeline.GetConfig()
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		logrus.Error("Must append policy id", r.URL.String())
		http.Error(w, "must append policy id", http.StatusBadRequest)
		return
	}

	logrus.Infof("Removing policy: %s", id)
	if id == global_name {
		conf.GlobalPolicy = &alarm.Policy{}
		conf.GlobalPolicy.Compile()
	} else {
		delete(conf.Policies, id)
	}

	conf.Provider().PutConfig(conf)
	p.pipeline.Refresh(conf)
}

// Post HTTP get method
func (p *PolicyConfig) Post(w http.ResponseWriter, r *http.Request) {
	conf := p.pipeline.GetConfig()
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		logrus.Error("Must append policy id", r.URL.String())
		http.Error(w, "must append policy id", http.StatusBadRequest)
		return
	}

	//  check to see if a policy with this id already exists
	if _, inMap := conf.Policies[id]; inMap {
		http.Error(w, fmt.Sprintf("A policy with id: '%s' already exists", id), http.StatusBadRequest)
		return
	}

	// read the policy
	buff, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pol := &alarm.Policy{}

	err = json.Unmarshal(buff, pol)
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if pol.Name == "" {
		pol.Name = id
	}

	pol.Compile()
	if conf.Policies == nil {
		conf.Policies = make(map[string]*alarm.Policy)
	}

	if id == global_name {
		conf.GlobalPolicy = pol
	} else {
		conf.Policies[id] = pol
	}

	conf.Provider().PutConfig(conf)

	p.pipeline.Refresh(conf)
}
