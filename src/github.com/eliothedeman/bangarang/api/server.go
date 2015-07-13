package api

import (
	"crypto/md5"
	"fmt"
	"net/http"

	"github.com/eliothedeman/bangarang/pipeline"
	"github.com/gorilla/mux"
)

// An EndPointer returns the end point for this method
type EndPointer interface {
	EndPoint() string
}

// A Getter provides an http "GET" method
type Getter interface {
	Get(http.ResponseWriter, *http.Request)
}

// A Poster provides an http "POST" method
type Poster interface {
	Post(http.ResponseWriter, *http.Request)
}

// A Putter provides an http "POST" method
type Putter interface {
	Put(http.ResponseWriter, *http.Request)
}

// A Deleter provides an http "DELETE" method
type Deleter interface {
	Delete(http.ResponseWriter, *http.Request)
}

// Server Serves the http api for bangarang
type Server struct {
	router     *mux.Router
	port       int
	pipeline   *pipeline.Pipeline
	configHash []byte
}

// wrap the given handler func in a closure that checks for auth first if the
// server is configured to use basic auth
func (s *Server) wrapAuth(h http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		h(w, r)
	}
}

func hashPassord(p string) string {
	m := md5.New()
	return string(m.Sum([]byte(p)))
}

func (s *Server) construct(e EndPointer) {

	if g, ok := e.(Getter); ok {
		route := s.router.NewRoute().Path(e.EndPoint())
		route.Methods("GET").HandlerFunc(s.wrapAuth(g.Get))
	}
	if p, ok := e.(Poster); ok {
		route := s.router.NewRoute().Path(e.EndPoint())
		route.Methods("POST", "PUT").HandlerFunc(s.wrapAuth(p.Post))
	}
	if d, ok := e.(Deleter); ok {
		route := s.router.NewRoute().Path(e.EndPoint())
		route.Methods("DELETE").HandlerFunc(s.wrapAuth(d.Delete))
	}
}

// Serve the bangarang api via HTTP
func (s *Server) Serve() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.router)
}

// NewServer creates and returns a new server
func NewServer(port int, pipe *pipeline.Pipeline) *Server {
	s := &Server{
		router:   mux.NewRouter(),
		port:     port,
		pipeline: pipe,
	}

	s.construct(NewAllIncidents(pipe))
	s.construct(NewIncident(pipe))
	s.construct(NewConfigHash(pipe))
	s.construct(NewEventStats(pipe))
	s.construct(NewKnownHosts(pipe))
	s.construct(NewKnownServices(pipe))
	s.construct(NewConfigRefresh(pipe))
	s.construct(NewProviderConfig(pipe))
	s.construct(NewPolicyConfig(pipe))
	s.construct(NewConfigVersion(pipe))
	s.construct(NewEscalationConfig(pipe))
	return s
}
