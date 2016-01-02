package api

import (
	"crypto/md5"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/pipeline"
	"github.com/gorilla/mux"
)

const (
	SESSION_HEADER_NAME   = "BANG_SESSION"
	EXPIRED_SESSION_TOKEN = 4000
	INVALID_SESSION_TOKEN = 4001
)

type Request struct {
	r *http.Request
	w http.ResponseWriter
	u *config.User
}

type RequestHandler func(*Request)

// An EndPointer returns the end point for this method
type EndPointer interface {
	EndPoint() string
}

// A Getter provides an http "GET" method
type Getter interface {
	Get(*Request)
}

type Get RequestHandler
type Post RequestHandler
type Put RequestHandler
type Delete RequestHandler

func (g Get) NeedsAuth() config.UserPermissions {
	return config.READ
}

// A Poster provides an http "POST" method
type Poster interface {
	Post(*Request)
}

func (p Post) NeedsAuth() config.UserPermissions {
	return config.WRITE
}

// A Putter provides an http "POST" method
type Putter interface {
	Put(*Request)
}

func (p Put) NeedsAuth() config.UserPermissions {
	return config.WRITE
}

// A Deleter provides an http "DELETE" method
type Deleter interface {
	Delete(*Request)
}

func (d Delete) NeedsAuth() config.UserPermissions {
	return config.WRITE
}

type NeedsAuther interface {
	NeedsAuth() config.UserPermissions
}

// Server Serves the http api for bangarang
type Server struct {
	router     *mux.Router
	port       int
	pipeline   *pipeline.Pipeline
	configHash []byte
}

func authUser(confProvider config.Provider, r *http.Request) (*config.User, error) {

	// check for a session token
	session := r.Header.Get(SESSION_HEADER_NAME)

	if session == "" {
		// attempt to get it form cookie
		c, _ := r.Cookie(SESSION_HEADER_NAME)
		if c != nil {
			session = c.Value
		}
	}

	// create user doesn't require auth
	if r.URL.Path == "/api/user" && r.Method == "POST" {
		return confProvider.GetUserByUserName("admin")
	}

	// fetch the user id from the session store for this token
	if session != "" {
		userName, err := GlobalSession.Get(session)
		if err != nil {
			return nil, err
		}

		//  get the user by the given id
		return confProvider.GetUser(userName)
	}
	user, password, ok := r.BasicAuth()
	if !ok {
		return nil, fmt.Errorf("Auth not provided")
	}

	// fetch the user
	u, err := confProvider.GetUserByUserName(user)
	if err != nil {
		return nil, err
	}

	// check to see if the password is correct
	if !config.CheckUserPassword(u, password) {
		return nil, fmt.Errorf("The provided password is incorrect for user %s", user)
	}

	return u, nil
}

// wrap the given handler func in a closure that checks for auth first if the
// server is configured to use basic auth
func (s *Server) wrapAuth(h interface{}) http.HandlerFunc {

	call := func(w http.ResponseWriter, r *http.Request) {

		var u *config.User
		var err error

		s.pipeline.ViewConfig(func(conf *config.AppConfig) {
			u, err = authUser(conf.Provider(), r)
		})

		req := &Request{
			u: u,
			r: r,
			w: w,
		}
		switch h := h.(type) {
		case Get:
			h(req)
		case Post:
			h(req)
		case Put:
			h(req)
		case Delete:
			h(req)
		}
	}

	// check to see if this auther needs
	if needsAuth, is := h.(NeedsAuther); is {
		needs := needsAuth.NeedsAuth()

		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("Access-Control-Allow-Origin", "*")
			var u *config.User
			var err error

			s.pipeline.ViewConfig(func(conf *config.AppConfig) {
				u, err = authUser(conf.Provider(), r)
			})

			// handle error types
			if err == InvalidSessionToken {
				http.Error(w, err.Error(), INVALID_SESSION_TOKEN)
				logrus.Error(err)
				return
			}

			if err == ExpiredSessionToken {
				http.Error(w, err.Error(), EXPIRED_SESSION_TOKEN)
				logrus.Error(err)
				return
			}

			if err != nil {
				if needs == config.READ {
					call(w, r)
					return
				}

				http.Error(w, err.Error(), http.StatusUnauthorized)
				logrus.Errorf("Permission Denied: %s", err.Error())
				return
			}

			// make sure we have at least the required permissions
			if u.Permissions < needs {
				http.Error(w, config.InsufficientPermissions(needs, u.Permissions).Error(), http.StatusForbidden)
				return
			}

			// if all is well, pass it on
			call(w, r)
		}
	}

	// else auth is not required
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Content-Type", "application/json")
		call(w, r)
	}
}

func hashPassord(p string) string {
	m := md5.New()
	return string(m.Sum([]byte(p)))
}

func (s *Server) construct(e EndPointer) {

	if g, ok := e.(Getter); ok {
		route := s.router.NewRoute().Path(e.EndPoint())
		route.Methods("GET").HandlerFunc(s.wrapAuth(Get(g.Get)))
	}
	if p, ok := e.(Poster); ok {
		route := s.router.NewRoute().Path(e.EndPoint())
		route.Methods("POST", "PUT").HandlerFunc(s.wrapAuth(Post(p.Post)))
	}
	if d, ok := e.(Deleter); ok {
		route := s.router.NewRoute().Path(e.EndPoint())
		route.Methods("DELETE").HandlerFunc(s.wrapAuth(Delete(d.Delete)))
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

	s.construct(NewIncident(pipe))
	s.construct(NewSystemStats(pipe))
	s.construct(NewConfigHash(pipe))
	s.construct(NewEventStats(pipe))
	s.construct(NewProviderConfig(pipe))
	s.construct(NewPolicyConfig(pipe))
	s.construct(NewConfigVersion(pipe))
	s.construct(NewEscalationConfig(pipe))
	s.construct(NewTag(pipe))
	s.construct(NewAuthUser(pipe))
	s.construct(NewUser(pipe))
	s.construct(NewUserPermissions(pipe))
	s.construct(NewUserPassword(pipe))
	return s
}
