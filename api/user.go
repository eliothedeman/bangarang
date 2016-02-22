package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/pipeline"
)

// User handles the api methods for incidents
type User struct {
	pipeline *pipeline.Pipeline
}

// NewUser Create a new User api method
func NewUser(pipe *pipeline.Pipeline) *User {
	return &User{
		pipeline: pipe,
	}
}

// EndPoint return the endpoint of this method
func (u *User) EndPoint() string {
	return "/api/user"
}

type GetUserResponse struct {
	UserName    string `json:"user_name"`
	Name        string `json:"name"`
	Permissions string `json:"permissions"`
}

func getUserResponseFromUser(u *config.User) *GetUserResponse {
	gur := &GetUserResponse{}
	gur.UserName = u.UserName
	gur.Name = u.Name
	gur.Permissions = config.PermissionsToName(u.Permissions)
	return gur
}

// Get fetches information about the spesified user
func (u *User) Get(req *Request) {
	uName := req.r.URL.Query().Get("user")

	u.pipeline.ViewConfig(func(conf *config.AppConfig) {
		var resp []*GetUserResponse
		// handle the "all" case
		if uName == "*" {
			users, err := conf.Provider().ListUsers()
			if err != nil {
				http.Error(req.w, err.Error(), http.StatusBadRequest)
				logrus.Error(err)
				return
			}
			resp = make([]*GetUserResponse, 0, len(users))
			for _, usr := range users {
				resp = append(resp, getUserResponseFromUser(usr))
			}
		} else if len(uName) > 0 {
			// handle the single case
			usr, err := conf.Provider().GetUserByUserName(uName)
			if err != nil {
				http.Error(req.w, err.Error(), http.StatusBadRequest)
				logrus.Error(err)
				return
			}
			resp = []*GetUserResponse{
				getUserResponseFromUser(usr),
			}
		} else {
			// handle the "get self" case
			usr, err := authUser(conf.Provider(), req.r)
			if err != nil {
				http.Error(req.w, err.Error(), http.StatusBadRequest)
				logrus.Error(err)
				return
			}

			resp = []*GetUserResponse{
				getUserResponseFromUser(usr),
			}

		}

		buff, err := json.Marshal(resp)
		if err != nil {
			http.Error(req.w, err.Error(), http.StatusInternalServerError)
			logrus.Error(err)
			return
		}

		req.w.Write(buff)
	})
}

type NewUserRequest struct {
	UserName string `json:"user_name"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// Post creates a new user
func (u *User) Post(req *Request) {
	buff, err := ioutil.ReadAll(req.r.Body)
	if err != nil {
		http.Error(req.w, err.Error(), http.StatusBadRequest)
		logrus.Error(err)
		return
	}
	nur := &NewUserRequest{}
	err = json.Unmarshal(buff, nur)
	if err != nil {
		http.Error(req.w, err.Error(), http.StatusBadRequest)
		logrus.Error(err)
		return
	}

	err = u.pipeline.UpdateConfig(func(conf *config.AppConfig) error {

		// create the user in  the database
		nu := config.NewUser(nur.Name, nur.UserName, nur.Password, config.READ)
		err = conf.Provider().PutUser(nu)
		return nil

	}, req.u)

	if err != nil {
		http.Error(req.w, err.Error(), http.StatusInternalServerError)
		logrus.Error(err)
		return
	}
}
