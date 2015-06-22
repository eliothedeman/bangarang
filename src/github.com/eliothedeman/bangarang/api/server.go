package api

import (
	"crypto/md5"
	"fmt"
	"net/http"

	"github.com/eliothedeman/bangarang/config"
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
	router      *mux.Router
	port        int
	pipeline    *pipeline.Pipeline
	auths       []config.BasicAuth
	config_hash []byte
}

// wrap the given handler func in a closure that checks for auth first if the server is configured to use basic auth
func (s *Server) wrapAuth(h http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")

		// if the server is configured to use basic auth, check the auth before proceeding to the request
		if len(s.auths) > 0 {
			u, p, ok := r.BasicAuth()
			if !ok {
				http.Error(w, "Basic auth must be supplied", http.StatusForbidden)
				return
			}

			if !s.authUser(u, p) {
				http.Error(w, "User or password are invalid", http.StatusForbidden)
				return
			}
		}

		// if the request passes auth, send it on
		h(w, r)
	}
}

func hashPassord(p string) string {
	m := md5.New()
	return string(m.Sum([]byte(p)))
}

func (s *Server) authUser(u, p string) bool {
	for _, a := range s.auths {
		if a.UserName == u {
			if a.PasswordHash == hashPassord(p) {
				return true
			}
		}
	}

	return false
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

func (s *Server) Serve() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.router)
}

func NewServer(port int, pipe *pipeline.Pipeline,
	auths []config.BasicAuth) *Server {
	s := &Server{
		router:   mux.NewRouter(),
		port:     port,
		pipeline: pipe,
		auths:    auths,
	}

	s.construct(NewAllIncidents(pipe))
	s.construct(NewIncident(pipe))
	s.construct(NewConfigHash(pipe))
	s.construct(NewEventStats(pipe))
	s.construct(NewKnownHosts(pipe))
	s.construct(NewKnownServices(pipe))
	s.construct(NewConfigRefresh(pipe))
	return s
}
