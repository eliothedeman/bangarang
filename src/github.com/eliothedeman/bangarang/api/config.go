package api

import (
	"encoding/json"
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
	Hash      string `json:"hash"`
	Timestamp int64  `json:"time"`
}

// Get HTTP get method
func (c *ConfigHash) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")

	snaps := c.pipeline.GetConfig().Provider().ListSnapshots()
	res := make([]*HashResponse, 0, len(snaps))

	for _, s := range snaps {
		res = append(res, &HashResponse{
			Hash:      s.Hash,
			Timestamp: s.Timestamp.Unix(),
		})
	}

	buf, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Write(buf)
}
