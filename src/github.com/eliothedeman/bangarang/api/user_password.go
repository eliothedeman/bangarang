package api

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/pipeline"
)

// UserPassword handles the api methods for incidents
type UserPassword struct {
	pipeline *pipeline.Pipeline
}

// NewUserPassword Create a new UserPassword api method
func NewUserPassword(pipe *pipeline.Pipeline) *UserPassword {
	return &UserPassword{
		pipeline: pipe,
	}
}

// EndPoint return the endpoint of this method
func (u *UserPassword) EndPoint() string {
	return "/api/user/password"
}

// Post changes the permissions for this user
func (up *UserPassword) Post(req *Request) {

	// MUST HAS ADMIN
	if req.u.Permissions != config.ADMIN {
		http.Error(req.w, config.InsufficientPermissions(config.ADMIN, req.u.Permissions).Error(), http.StatusBadRequest)
		return
	}

	// get the user we want to update
	q := req.r.URL.Query()
	userName := q.Get("user")
	if userName == "" {
		http.Error(req.w, "user name must be supplied", http.StatusBadRequest)
		return
	}

	newPass := q.Get("new")
	if newPass == "" {
		http.Error(req.w, "new password must be supplied", http.StatusBadRequest)
		return
	}

	err := up.pipeline.UpdateConfig(func(conf *config.AppConfig) error {
		userToUpdate, err := conf.Provider().GetUser(userName)
		if err != nil {
			return err
		}

		userToUpdate.PasswordHash = config.HashUserPassword(userToUpdate, newPass)
		return conf.Provider().PutUser(userToUpdate)

	}, req.u)
	if err != nil {
		http.Error(req.w, err.Error(), http.StatusInternalServerError)
		logrus.Error(err)
		return
	}

	// success
}
