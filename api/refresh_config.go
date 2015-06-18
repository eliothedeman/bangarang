package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/pipeline"
)

// refresh the config file, and restart the server
type ConfigRefresh struct {
	pipeline *pipeline.Pipeline
}

func NewConfigRefresh(pipe *pipeline.Pipeline) *ConfigRefresh {
	return &ConfigRefresh{
		pipeline: pipe,
	}
}

func (i *ConfigRefresh) EndPoint() string {
	return "/api/config/refresh"
}

func (c *ConfigRefresh) Get(w http.ResponseWriter, r *http.Request) {

	// if there is no path given, reload the file currently configed
	path := r.URL.Query().Get("path")
	if path == "" {
		path = c.pipeline.GetConfig().FileName()
	}

	// reload the config for the pipeline
	conf, err := config.LoadConfigFile(path)
	if err != nil {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logrus.Error(err)
			return
		}
	}

	c.pipeline.Refresh(conf)

	w.Header().Add("content-type", "application/json")

	res := &HashResponse{
		Hash: fmt.Sprintf("%x", c.pipeline.GetConfig().Hash),
	}

	buf, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logrus.Error(err)
		return
	}

	w.Write(buf)
}
