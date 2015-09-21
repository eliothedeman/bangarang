package api

import (
	"encoding/json"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/pipeline"
)

// AuthUser handles the api methods for incidents
type AuthUser struct {
	pipeline *pipeline.Pipeline
}

// NewAuthUser Create a new AuthUser api method
func NewAuthUser(pipe *pipeline.Pipeline) *AuthUser {
	return &AuthUser{
		pipeline: pipe,
	}
}

// EndPoint return the endpoint of this method
func (c *AuthUser) EndPoint() string {
	return "/api/auth/user"
}

// Get HTTP get method
func (c *AuthUser) Get(w http.ResponseWriter, r *http.Request) {

	// get the username/password
	user := r.URL.Query().Get("user")
	pass := r.URL.Query().Get("pass")

	// fetch the user form the db
	var u *config.User
	var err error
	c.pipeline.ViewConfig(func(ac *config.AppConfig) {
		u, err = ac.Provider().GetUserByUserName(user)
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logrus.Error(err)
		return
	}

	// if the user does not have the correct password, fail
	if !config.CheckUserPassword(u, pass) {
		http.Error(w, "invalid password", http.StatusBadRequest)
		return
	}

	token := GlobalSession.Put(u.UserName)

	// encode the response as json
	buff, err := json.Marshal(map[string]string{
		"token": token,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logrus.Error(err)
		return
	}

	w.Header().Add("content-type", "application/json")
	w.Write(buff)
}
