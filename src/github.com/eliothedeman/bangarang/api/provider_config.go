package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/config"
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
func (c *ProviderConfig) Get(req *Request) {
	c.pipeline.ViewConfig(func(cfg *config.AppConfig) {
		confs := cfg.EventProviders.Raw()
		vars := mux.Vars(req.r)
		id, ok := vars["id"]
		if !ok {
			http.Error(req.w, "must append provider id", http.StatusBadRequest)
			return
		}

		// if the provider is "*" fetch all configs
		if id == "*" {
			buff, err := json.Marshal(confs)
			if err != nil {
				logrus.Error(err)
				http.Error(req.w, err.Error(), http.StatusInternalServerError)
			}
			req.w.Write(buff)
			return
		}

		conf, ok := confs[id]
		if !ok {
			http.Error(req.w, fmt.Sprintf("Unknown event provider %s", id), http.StatusBadRequest)
			return
		}

		req.w.Write(conf)
	})
}

// Delete the given event provider
func (p *ProviderConfig) Delete(req *Request) {
	err := p.pipeline.UpdateConfig(func(conf *config.AppConfig) error {
		vars := mux.Vars(req.r)
		id, ok := vars["id"]
		if !ok {
			return fmt.Errorf("Must append provider id %s", req.r.URL)
		}

		delete(conf.EventProviders.Collection, id)
		delete(conf.EventProviders.Raw(), id)

		return nil

	}, req.u)

	if err != nil {
		logrus.Error(err)
		http.Error(req.w, err.Error(), http.StatusBadRequest)
	}

}

// Post HTTP get method
func (c *ProviderConfig) Post(req *Request) {

	// get the config.Provider for our current config
	err := c.pipeline.UpdateConfig(func(conf *config.AppConfig) error {
		vars := mux.Vars(req.r)
		id, ok := vars["id"]
		if !ok {
			return fmt.Errorf("Must append provider id %s", req.r.URL)
		}
		if _, inMap := conf.EventProviders.Collection[id]; inMap {
			return fmt.Errorf("Provider \"%s\" already exists", id)
		}

		// read out the new raw provider
		buff, err := ioutil.ReadAll(req.r.Body)
		if err != nil {
			return err
		}

		ep, err := provider.ParseProvider(buff)
		if err != nil {
			return err
		}

		if conf.EventProviders.Collection == nil {
			conf.EventProviders.Collection = make(map[string]provider.EventProvider)
		}

		logrus.Infof("Adding new event provider %s", id)
		conf.EventProviders.Add(id, ep, buff)
		return nil

	}, req.u)

	if err != nil {
		logrus.Error(err)
		http.Error(req.w, err.Error(), http.StatusBadRequest)
		return
	}

}
