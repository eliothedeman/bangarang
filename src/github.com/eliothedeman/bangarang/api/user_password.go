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
func (up *UserPassword) Post(w http.ResponseWriter, r *http.Request) {
	u, err := authUser(up.pipeline.GetConfig().Provider(), r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logrus.Error(err)
		return
	}

	// MUST HAS ADMIN
	if u.Permissions != config.ADMIN {
		http.Error(w, config.InsufficientPermissions(config.ADMIN, u.Permissions).Error(), http.StatusBadRequest)
		return
	}

	// get the user we want to update
	q := r.URL.Query()
	userName := q.Get("user")
	if userName == "" {
		http.Error(w, "user name must be supplied", http.StatusBadRequest)
		return
	}

	newPass := q.Get("new")
	if newPass == "" {
		http.Error(w, "new password must be supplied", http.StatusBadRequest)
		return
	}

	userToUpdate, err := up.pipeline.GetConfig().Provider().GetUser(userName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logrus.Error(err)
		return
	}

	// can't update admin's permissions
	if userToUpdate.UserName == "admin" {
		http.Error(w, "updating admin's permissions is not allowed", http.StatusBadRequest)
		return
	}

	userToUpdate.PasswordHash = config.HashUserPassword(userToUpdate, newPass)
	err = up.pipeline.GetConfig().Provider().PutUser(userToUpdate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logrus.Error(err)
		return
	}

	// success
}
