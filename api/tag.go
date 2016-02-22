package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/eliothedeman/bangarang/event"
	"github.com/eliothedeman/bangarang/pipeline"
	"github.com/gorilla/mux"
)

// handles the api methods for incidents
type Tag struct {
	pipeline *pipeline.Pipeline
}

func NewTag(p *pipeline.Pipeline) *Tag {
	return &Tag{
		pipeline: p,
	}
}

func (t *Tag) EndPoint() string {
	return "/api/tag/{key:.+}/{value:.+}"
}

func GetTag(w http.ResponseWriter, r *http.Request) (kv event.KeyVal) {
	vars := mux.Vars(r)
	key, ok := vars["key"]
	if !ok {
		http.Error(w, "Must append a tag key", http.StatusBadRequest)
		return
	}

	value, ok := vars["value"]
	if !ok {
		http.Error(w, "Must append a tag value", http.StatusBadRequest)
		return
	}

	kv.Key = key
	kv.Value = value

	return
}

// Delete will resolve a given event
func (t *Tag) Delete(w http.ResponseWriter, r *http.Request) {
	kv := GetTag(w, r)
	t.pipeline.GetTracker().RemoveTag(kv.Key, kv.Value)
}

func (t *Tag) Get(w http.ResponseWriter, r *http.Request) {
	kv := GetTag(w, r)
	tkr := t.pipeline.GetTracker()
	var tags []string

	// get all elements of this tag
	if kv.Value == "*" {
		tags = tkr.GetTag(kv.Key)

		// get all tags that have this prefix
	} else {
		tmp := tkr.GetTag(kv.Key)
		tags = make([]string, 0, len(tmp))

		// search through an only include the tags values that match the given prefix
		for _, tag := range tags {
			if strings.HasPrefix(tag, kv.Value) {
				tags = append(tags, tag)
			}
		}
	}

	// encode and return the list
	buff, err := json.Marshal(tags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(buff)
}
