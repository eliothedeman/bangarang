package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/eliothedeman/bangarang/pipeline"
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

	conf, ok := confs[id]
	if !ok {
		http.Error(w, fmt.Sprintf("Unknown event provider %s", id), http.StatusBadRequest)
		return
	}

	w.Write(conf)
}

// Post HTTP get method
func (c *ProviderConfig) Post(w http.ResponseWriter, r *http.Request) {

	// get the config.Provider for our current config
	conf := c.pipeline.GetConfig()
	p := conf.Provider()

	// read out the new raw provider
	buff, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "must append provider id", http.StatusBadRequest)
		return
	}

}
