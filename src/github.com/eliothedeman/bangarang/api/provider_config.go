package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/pipeline"
	"github.com/eliothedeman/bangarang/provider"
	"github.com/gorilla/mux"
)

// ProviderConfig handles the api methods for incidents
type ProviderConfig struct {
	pipeline *pipeline.Pipeline
}

func NewProviderConfig(pipe *pipeline.Pipeline) *ProviderConfig {
	return &ProviderConfig{
		pipeline: pipe,
	}
}

// EndPoint return the endpoint of this method
func (c *ProviderConfig) EndPoint() string {
	return "/api/provider/config/{id}"
}

// Get HTTP get method
func (c *ProviderConfig) Get(w http.ResponseWriter, r *http.Request) {
	confs := c.pipeline.GetConfig().EventProviders.Raw()
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		http.Error(w, "must append provider id", http.StatusBadRequest)
		return
	}

	// if the provider is "*" fetch all configs
	if id == "*" {
		buff, err := json.Marshal(confs)
		if err != nil {
			logrus.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write(buff)
		return
	}

	conf, ok := confs[id]
	if !ok {
		http.Error(w, fmt.Sprintf("Unknown event provider %s", id), http.StatusBadRequest)
		return
	}

	w.Write(conf)
}

// Delete the given event provider
func (p *ProviderConfig) Delete(w http.ResponseWriter, r *http.Request) {
	// get the user for this method
	u, err := authUser(p.pipeline.GetConfig().Provider(), r)
	if err != nil {
		if err != nil {
			logrus.Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	conf := p.pipeline.GetConfig()
	cp := conf.Provider()
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		logrus.Error("Must append provider id", r.URL.String())
		http.Error(w, "must append provider id", http.StatusBadRequest)
		return
	}

	delete(conf.EventProviders.Collection, id)
	delete(conf.EventProviders.Raw(), id)

	// refresh the config without the provider
	cp.PutConfig(conf, u)
	p.pipeline.Refresh(conf)
}

// Post HTTP get method
func (c *ProviderConfig) Post(w http.ResponseWriter, r *http.Request) {
	// get the user for this method
	u, err := authUser(c.pipeline.GetConfig().Provider(), r)
	if err != nil {
		if err != nil {
			logrus.Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// get the config.Provider for our current config
	conf := c.pipeline.GetConfig()
	p := conf.Provider()
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		logrus.Error("Must append provider id", r.URL.String())
		http.Error(w, "must append provider id", http.StatusBadRequest)
		return
	}
	if _, inMap := conf.EventProviders.Collection[id]; inMap {
		http.Error(w, fmt.Sprintf("Provider \"%s\" already exists", id), http.StatusBadRequest)
		return
	}

	// read out the new raw provider
	buff, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ep, err := provider.ParseProvider(buff)
	if err != nil {
		if err != nil {
			logrus.Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if conf.EventProviders.Collection == nil {
		conf.EventProviders.Collection = make(map[string]provider.EventProvider)
	}

	conf.EventProviders.Add(id, ep, buff)

	// write the new config
	_, err = p.PutConfig(conf, u)
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// pull the new config out, and restart the pipeline
	conf, err = conf.Provider().GetCurrent()
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.pipeline.Refresh(conf)
}
