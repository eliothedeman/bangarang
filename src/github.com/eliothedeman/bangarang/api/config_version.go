package api

import (
	"encoding/json"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/pipeline"
	"github.com/gorilla/mux"
)

// ConfigVersion handles the api methods for incidents
type ConfigVersion struct {
	pipeline *pipeline.Pipeline
}

// NewConfigVersion Create a new ConfigVersion api method
func NewConfigVersion(pipe *pipeline.Pipeline) *ConfigVersion {
	return &ConfigVersion{
		pipeline: pipe,
	}
}

// EndPoint return the endpoint of this method
func (c *ConfigVersion) EndPoint() string {
	return "/api/config/version/{version}"
}

// Get HTTP get method
func (c *ConfigVersion) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	vars := mux.Vars(r)
	ver, ok := vars["version"]
	if !ok {
		http.Error(w, "must append config version", http.StatusBadRequest)
		return
	}

	p := c.pipeline.GetConfig().Provider()

	// return all config versions
	if ver == "*" {
		buff, err := json.Marshal(p.ListRawSnapshots())
		if err != nil {
			logrus.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(buff)
		return
	}

	conf, err := p.GetConfig(vars["version"])
	if err != nil {
		if err != nil {
			logrus.Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	buff, err := json.Marshal(conf)
	if err != nil {
		if err != nil {
			logrus.Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	w.Write(buff)
}

// change the current config to a spesific version
func (c *ConfigVersion) Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")

	// get the user for this method
	u, err := authUser(c.pipeline.GetConfig().Provider(), r)
	if err != nil {
		if err != nil {
			logrus.Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	vars := mux.Vars(r)
	version, ok := vars["version"]
	if !ok {
		http.Error(w, "must append config version", http.StatusBadRequest)
		return
	}

	p := c.pipeline.GetConfig().Provider()

	// get the config that this version is looking for
	conf, err := p.GetConfig(version)
	if err != nil {
		if err != nil {
			logrus.Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	_, err = p.PutConfig(conf, u)
	if err != nil {
		if err != nil {
			logrus.Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	c.pipeline.Refresh(conf)

}
