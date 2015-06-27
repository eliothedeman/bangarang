package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/eliothedeman/bangarang/pipeline"
)

// ConfigHash handles the api methods for incidents
type ConfigHash struct {
	pipeline *pipeline.Pipeline
}

// NewConfigHash Create a new ConfigHash api method
func NewConfigHash(pipe *pipeline.Pipeline) *ConfigHash {
	return &ConfigHash{
		pipeline: pipe,
	}
}

// EndPoint return the endpoint of this method
func (c *ConfigHash) EndPoint() string {
	return "/api/config/hash"
}

// HashResponse will be marshaled as json as the successful response
type HashResponse struct {
	Hash string `json:"hash"`
}

// Get HTTP get method
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
	w.Header()

	w.Write(buf)
}
