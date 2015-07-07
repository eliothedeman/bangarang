package main

import (
	"flag"
	"io"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
)

type Server struct {
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	path := r.URL.Path

	if strings.HasPrefix(path, "/api") {
		logrus.Infof("Relaying api call %s", path)
		r.RequestURI = ""
		r.URL.Scheme = "http"
		r.URL.Host = *api_host
		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			logrus.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			logrus.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
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

var (
	api_host = flag.String("api", "localhost:8081", "")
	listen   = flag.String("l", ":9090", "serve http on")
	root     = flag.String("r", ".", "root dir for assets")
	rootDir  = "."
)

func main() {
	flag.Parse()
	rootDir = *root
	s := &Server{}

	http.ListenAndServe(*listen, s)
}
