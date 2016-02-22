package api

import (
	"encoding/json"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/config"
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
func (c *ConfigVersion) Get(req *Request) {
	vars := mux.Vars(req.r)
	ver, ok := vars["version"]
	if !ok {
		http.Error(req.w, "must append config version", http.StatusBadRequest)
		return
	}

	var conf *config.AppConfig
	c.pipeline.ViewConfig(func(x *config.AppConfig) {
		conf = x
	})

	p := conf.Provider()

	// return all config versions
	if ver == "*" {
		buff, err := json.Marshal(p.ListSnapshots())
		if err != nil {
			logrus.Error(err)
			http.Error(req.w, err.Error(), http.StatusInternalServerError)
			return
		}
		req.w.Write(buff)
		return
	}

	conf, err := p.GetConfig(vars["version"])
	if err != nil {
		logrus.Error(err)
		http.Error(req.w, err.Error(), http.StatusBadRequest)
		return
	}

	buff, err := json.Marshal(conf)
	if err != nil {
		logrus.Error(err)
		http.Error(req.w, err.Error(), http.StatusBadRequest)
		return
	}

	req.w.Write(buff)
}

// change the current config to a spesific version
func (c *ConfigVersion) Post(req *Request) {

	var p config.Provider
	c.pipeline.ViewConfig(func(conf *config.AppConfig) {
		p = conf.Provider()
	})

	vars := mux.Vars(req.r)
	version, ok := vars["version"]
	if !ok {
		http.Error(req.w, "must append config version", http.StatusBadRequest)
		return
	}

	// get the config that this version is looking for
	conf, err := p.GetConfig(version)
	if err != nil {
		logrus.Error(err)
		http.Error(req.w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = p.PutConfig(conf, req.u)
	if err != nil {
		logrus.Error(err)
		http.Error(req.w, err.Error(), http.StatusBadRequest)
		return
	}

	c.pipeline.Refresh(conf)

}
