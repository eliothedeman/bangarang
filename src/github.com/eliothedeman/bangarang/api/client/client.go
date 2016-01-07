package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/event"
)

type Client struct {
	host    string
	port    int
	headers http.Header
}

func NewClient() *Client {
	return &Client{
		host:    "localhost",
		port:    8081,
		headers: make(http.Header),
	}
}

func NewClientWithAuthToken(token string) *Client {
	c := NewClient()

	// only set the token if it is valid
	if len(token) != 0 {
		c.headers.Add("BANG_SESSION", token)
	}
	return c
}

// Do is a helper function for making http requests to the api
func (c *Client) Do(url, method string, body, resp interface{}) error {
	b := bytes.NewBuffer(nil)

	// make a proper url if one is not given
	if url[0] != 'h' {
		url = c.GenUrl(url, nil)
	}

	if body != nil {
		buff, err := json.Marshal(body)
		if err != nil {
			return err
		}

		b.Write(buff)
	}

	r, err := http.NewRequest(method, url, b)
	if err != nil {
		return err
	}
	r.Header = c.headers

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}

	buff, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(buff, resp)

	return err
}

func (c *Client) GenUrl(path string, params map[string]string) string {
	if path[0] == '/' {
		path = path[1:]
	}

	base := fmt.Sprintf("http://%s:%d/%s", c.host, c.port, path)
	if params == nil {
		return base
	}
	base += "?"

	var x, y int
	x = len(params)
	for k, v := range params {
		fmt.Sprintf("%s=%s", k, v)
		y++
		if x != y {
			base += fmt.Sprintf("%s=%s&", k, v)
		} else {
			base += fmt.Sprintf("%s=%s", k, v)
		}
	}

	return base
}

func (c *Client) AuthUser(username, password string) (string, error) {
	m := map[string]string{}
	err := c.Do(c.GenUrl("api/auth/user", map[string]string{"user": username, "pass": password}), "get", nil, &m)
	if err != nil {
		return "", err
	}

	token, ok := m["token"]
	if !ok {
		return "", errors.New("Bad response")
	}

	return token, nil
}

// GetSelf should only be called if the correct headers are set
func (c *Client) GetSelf() (*config.User, error) {
	u := []*config.User{}
	err := c.Do(c.GenUrl("api/user", nil), "GET", nil, &u)

	if len(u) == 0 {
		return nil, err
	}
	return u[0], err
}

func (c *Client) GetIncidents(offset, max int) ([]*event.Incident, error) {
	mi := map[string]*event.Incident{}
	err := c.Do("api/incident/*", "GET", nil, &mi)
	if err != nil {
		return []*event.Incident{}, err
	}

	i := make([]*event.Incident, 0, len(mi))

	for _, v := range mi {
		i = append(i, v)
	}

	if offset > len(i) {
		return []*event.Incident{}, nil
	}

	if max < 0 {
		max = 0
	}

	if offset < 0 {
		offset = 0
	}

	if offset+max > len(i) {
		max = offset - len(i)
	}

	if max == 0 {
		return i[offset:], nil
	}

	return i[offset : offset+max], nil
}
