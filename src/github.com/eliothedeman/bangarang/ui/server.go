package ui

import (
	"net/http"

	"github.com/Sirupsen/logrus"
)

type Server struct {
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	path := r.URL.Path

	// serve the index.html if no path is given
	if path == "/" || path == "" {
		path = "index.html"
	} else {
		path = path[1:]
	}

	buff, err := Asset(path)
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Write(buff)
}
