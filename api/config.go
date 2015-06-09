package api

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// handles the api methods for incidents
type ConfigInfo struct {
	config_hash []byte
}

func NewConfig(config_hash []byte) *ConfigInfo {
	return &ConfigInfo{
		config_hash: config_hash,
	}
}

func (i *ConfigInfo) EndPoint() string {
	return "/api/config/hash"
}

type HashResponse struct {
	Hash string `json:"hash"`
}

func (c *ConfigInfo) Get(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("content-type", "application/json")

	res := &HashResponse{
		Hash: fmt.Sprintf("%x", md5.Sum(c.config_hash)),
	}

	buf, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Write(buf)
}
