package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/config"
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
func (p *PolicyConfig) Get(req *Request) {
	p.pipeline.ViewConfig(func(conf *config.AppConfig) {
		vars := mux.Vars(req.r)
		id, ok := vars["id"]
		if !ok {
			logrus.Error("Must append policy id", req.r.URL.String())
			http.Error(req.w, "must append policy id", http.StatusBadRequest)
			return
		}

		if id == "*" {
			buff, err := json.Marshal(conf.Policies)
			if err != nil {
				logrus.Error(err)
				http.Error(req.w, err.Error(), http.StatusBadRequest)
				return
			}

			req.w.Write(buff)
			return
		}

		// special case for global
		var pol *alarm.Policy
		if id == "global" {
			pol = conf.GlobalPolicy
		} else {
			pol, ok = conf.Policies[id]
			if !ok {
				http.Error(req.w, fmt.Sprintf("Unable to find policy '%s'", id), http.StatusBadRequest)
				return
			}
		}
		buff, err := json.Marshal(pol)
		if err != nil {
			logrus.Error(err)
			http.Error(req.w, err.Error(), http.StatusBadRequest)
			return
		}

		req.w.Write(buff)

	})
}

// Delete the given event provider
func (p *PolicyConfig) Delete(req *Request) {
	err := p.pipeline.UpdateConfig(func(conf *config.AppConfig) error {
		vars := mux.Vars(req.r)
		id, ok := vars["id"]
		if !ok {
			return fmt.Errorf("Must append policy id %s", req.r.URL)
		}

		logrus.Infof("Removing policy: %s", id)
		if id == global_name {
			conf.GlobalPolicy = &alarm.Policy{}
			conf.GlobalPolicy.Compile()
		} else {
			delete(conf.Policies, id)
			p.pipeline.RemovePolicy(id)
		}

		return nil

	}, req.u)

	if err != nil {
		logrus.Error(err)
		http.Error(req.w, err.Error(), http.StatusBadRequest)
	}

}

// Post HTTP get method
func (p *PolicyConfig) Post(req *Request) {
	// get the user for this method
	err := p.pipeline.UpdateConfig(func(conf *config.AppConfig) error {
		vars := mux.Vars(req.r)
		id, ok := vars["id"]
		if !ok {
			return fmt.Errorf("Must append policy id %s", req.r.URL)
		}

		//  check to see if a policy with this id already exists
		if _, inMap := conf.Policies[id]; inMap {

			return fmt.Errorf("A policy with id: '%s' already exists", id)
		}

		// read the policy
		buff, err := ioutil.ReadAll(req.r.Body)

		pol := &alarm.Policy{}

		err = json.Unmarshal(buff, pol)
		if err != nil {
			return err
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

		return nil

	}, req.u)

	if err != nil {
		logrus.Error(err)
		http.Error(req.w, err.Error(), http.StatusBadRequest)
		return
	}

}
