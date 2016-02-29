package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/eliothedeman/bangarang/api"
)

const (
	SessionTokenHeaderName = "BANG_SESSION"
	ClientTypeHeaderName   = "CLIENT_TYPE"
	GoClientHeaderValue    = "GO-OFFICIAL"
	GetMethod              = "GET"
	PostMethod             = "POST"
	DeleteMthod            = "DELETE"
)

var (
	ClientNotAuthenticated = errors.New("the client has not been authenticated")
)

func makeQueryString(m map[string]interface{}) string {
	b := bytes.NewBuffer(nil)
	if len(m) == 1 {
		for k, v := range m {
			b.WriteString(fmt.Sprintf("%s=%v", k, v))
		}
		return b.String()
	}

	for k, v := range m {
		b.WriteString(fmt.Sprintf("%s=%v", k, v))
		b.WriteString("&")
	}

	return b.String()
}

// An APIClient gives access to the http api of a bangarang instance
type APIClient struct {
	host         string
	port         int
	sessionToken string
}

// NewAPIClient creates and returns a new APIClient.
// The client will attemp to authenticate
func NewAPIClient(host string, port int, username, password string) (*APIClient, error) {
	a := &APIClient{
		host: host,
		port: port,
	}

	token, err := a.AuthUser(username, password)
	a.sessionToken = token
	return a, err
}

// NewPreAuthedClient returns a new client given a session token that has already been authed
func NewPreAuthedClient(host string, port int, token string) *APIClient {
	return &APIClient{
		host:         host,
		port:         port,
		sessionToken: token,
	}
}

func (a *APIClient) fmtHost() string {
	return fmtHost(a.host, a.port)
}

// buildRequest constructs the http request for a given method
func (a *APIClient) buildRequest(method, path, query string, body io.Reader) (*http.Request, error) {

	// make sure we have been authed
	if a.sessionToken == "" {
		return nil, ClientNotAuthenticated
	}

	// construct the request
	var url string
	if query != "" {
		url = fmt.Sprintf("%s/%s?%s", a.fmtHost(), path, query)
	} else {
		url = fmt.Sprintf("%s/%s", a.fmtHost(), path)
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// headers
	req.Header.Add(SessionTokenHeaderName, a.sessionToken)
	req.Header.Add(ClientTypeHeaderName, GoClientHeaderValue)

	return req, nil
}

func (a *APIClient) makeRequest(method, path, query string, body interface{}) ([]byte, error) {
	buff := bytes.NewBuffer(nil)
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buff.Write(b)
	}

	req, err := a.buildRequest(method, path, query, buff)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}

// AuthUser authorizes api access for a user and returns a session token
func (a *APIClient) AuthUser(username, password string) (string, error) {
	q := makeQueryString(map[string]interface{}{
		"user": username,
		"pass": password,
	})
	auth := &api.AuthUser{}

	buff, err := a.makeRequest(GetMethod, auth.EndPoint(), q, nil)
	if err != nil {
		return "", err
	}

	resp := map[string]string{}
	err = json.Unmarshal(buff, &resp)
	if err != nil {
		return "", err
	}

	return resp["token"], nil
}
