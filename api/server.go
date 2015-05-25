package api

import (
	"fmt"
	"net/http"

	"github.com/eliothedeman/bangarang/pipeline"
	"github.com/gorilla/mux"
)

// return the end point for this method
type EndPointer interface {
	EndPoint() string
}

// provides an http "GET" method
type Getter interface {
	Get(http.ResponseWriter, *http.Request)
}

// provides an http "POST" method
type Poster interface {
	Post(http.ResponseWriter, *http.Request)
}

// provides an http "DELETE" method
type Deleter interface {
	Delete(http.ResponseWriter, *http.Request)
}

// Serves the http api for bangarang
type Server struct {
	router   *mux.Router
	port     int
	pipeline *pipeline.Pipeline
}

func (s *Server) construct(e EndPointer) {
	route := s.router.NewRoute().Path(e.EndPoint())

	if g, ok := e.(Getter); ok {
		route.Methods("GET").HandlerFunc(g.Get)
	}
	if p, ok := e.(Poster); ok {
		route.Methods("POST", "PUT").HandlerFunc(p.Post)
	}
	if d, ok := e.(Deleter); ok {
		route.Methods("DELETE").HandlerFunc(d.Delete)
	}
}

func (s *Server) Serve() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.router)
}

func NewServer(port int, pipe *pipeline.Pipeline) *Server {
	s := &Server{
		router:   mux.NewRouter(),
		port:     port,
		pipeline: pipe,
	}

	s.construct(NewIncidents(pipe))

	return s
}
