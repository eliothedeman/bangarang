package api

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/pipeline"
)

var testPort = 8080
var configHash = []byte{0}

func newTestServerNoAuth() (*Server, int) {
	c := config.NewDefaultConfig()
	c.DbPath = fmt.Sprintf("%d.db", time.Now().UnixNano())
	p := pipeline.NewPipeline(c)
	testPort += 1
	s := NewServer(testPort, p, nil, configHash)
	return s, testPort
}

func newTestServerWithAuth(auths []config.BasicAuth) (*Server, int) {
	c := config.NewDefaultConfig()
	c.DbPath = fmt.Sprintf("%d.db", time.Now().UnixNano())
	p := pipeline.NewPipeline(c)
	testPort += 1
	s := NewServer(testPort, p, auths, nil)
	return s, testPort
}

func TestConfigHash(t *testing.T) {
	s, port := newTestServerNoAuth()
	go s.Serve()

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/api/config/hash", port))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var config_res HashResponse
	err = json.Unmarshal(body, &config_res)
	if err != nil {
		t.Fatal(err)
	}

	exp_hash := fmt.Sprintf("%x", md5.Sum(configHash))

	if exp_hash != config_res.Hash {
		t.Fatal(nil)
	}
}

func TestBasicAuth(t *testing.T) {
	s, port := newTestServerNoAuth()
	go s.Serve()

	_, err := http.Get(fmt.Sprintf("http://localhost:%d/api/all-incidents", port))
	if err != nil {
		t.Fatal(err)
	}

	s, port = newTestServerWithAuth([]config.BasicAuth{
		config.BasicAuth{
			UserName:     "test1",
			PasswordHash: hashPassord("password"),
		},
	})

	go s.Serve()
	resp, _ := http.Get(fmt.Sprintf("http://localhost:%d/api/all-incidents", port))

	if resp.StatusCode != http.StatusForbidden {
		t.Fatal("Should require auth")
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/api/all-incidents", port), nil)
	req.SetBasicAuth("test1", "password")

	resp, _ = http.DefaultClient.Do(req)

	if resp.StatusCode == http.StatusForbidden {
		t.Fatal("Should be allowed")
	}
}
