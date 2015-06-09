package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/eliothedeman/bangarang/pipeline"
)

// handles the api methods for incidents
type ConfigHash struct {
	pipeline *pipeline.Pipeline
}

func NewConfigHash(pipe *pipeline.Pipeline) *ConfigHash {
	return &ConfigHash{
		pipeline: pipe,
	}
}

func (i *ConfigHash) EndPoint() string {
	return "/api/config/hash"
}

type HashResponse struct {
	Hash string `json:"hash"`
}

func (c *ConfigHash) Get(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("content-type", "application/json")

	res := &HashResponse{
		Hash: fmt.Sprintf("%x", c.pipeline.GetConfig().Hash),
	}

	buf, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Write(buf)
}
