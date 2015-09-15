package api

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/pipeline"
	"github.com/gorilla/mux"
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
	return "/api/user/permissions/{user_name:.*}/{perms:.*}"
}

// Get all users with the given permissions
func (up *UserPermissions) Get(w http.ResponseWriter, r *http.Request) {

}

// Post changes the permissions for this user
func (up *UserPermissions) Post(w http.ResponseWriter, r *http.Request) {
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
	vars := mux.Vars(r)
	userName, ok := vars["user_name"]
	if !ok {
		http.Error(w, "user name must be supplied", http.StatusBadRequest)
		return
	}

	perms, ok := vars["perms"]
	if !ok {
		http.Error(w, "perms must be supplied", http.StatusBadRequest)
		return
	}

	uPerms := config.NameToPermissions(perms)
	if uPerms == -1 {
		if !ok {
			http.Error(w, fmt.Sprintf("invalid permissions %s", perms), http.StatusBadRequest)
			return
		}
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

	userToUpdate.Permissions = config.ADMIN
	err = up.pipeline.GetConfig().Provider().PutUser(userToUpdate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logrus.Error(err)
		return
	}

	// success
}
