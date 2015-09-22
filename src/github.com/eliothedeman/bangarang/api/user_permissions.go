package api

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/pipeline"
)

// UserPermissions handles the api methods for incidents
type UserPermissions struct {
	pipeline *pipeline.Pipeline
}

// NewUserPermissions Create a new UserPermissions api method
func NewUserPermissions(pipe *pipeline.Pipeline) *UserPermissions {
	return &UserPermissions{
		pipeline: pipe,
	}
}

// EndPoint return the endpoint of this method
func (u *UserPermissions) EndPoint() string {
	return "/api/user/permissions"
}

// Post changes the permissions for this user
func (up *UserPermissions) Post(req *Request) {

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

	perms := q.Get("perms")
	if perms == "" {
		http.Error(req.w, "perms must be supplied", http.StatusBadRequest)
		return
	}

	uPerms := config.NameToPermissions(perms)
	if uPerms == -1 {
		http.Error(req.w, fmt.Sprintf("invalid permissions %s", perms), http.StatusBadRequest)
		return
	}

	err := up.pipeline.UpdateConfig(func(conf *config.AppConfig) error {
		userToUpdate, err := conf.Provider().GetUser(userName)
		if err != nil {
			return err
		}

		// can't update admin's permissions
		if userToUpdate.UserName == "admin" {
			return fmt.Errorf("updating admin's permissions is not allowed")
		}

		userToUpdate.Permissions = uPerms
		return conf.Provider().PutUser(userToUpdate)

	}, req.u)
	if err != nil {
		http.Error(req.w, err.Error(), http.StatusInternalServerError)
		logrus.Error(err)
		return
	}

	// success
}
