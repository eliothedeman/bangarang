package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/pipeline"
)

var testPort = 8080

func newTestServerNoAuth() (*Server, int) {
	c := config.NewDefaultConfig()
	c.DbPath = fmt.Sprintf("%d.db", time.Now().UnixNano())
	p := pipeline.NewPipeline(c)
	testPort += 1
	s := NewServer(testPort, p, nil)
	return s, testPort
}

func newTestServerWithAuth(auths []config.BasicAuth) (*Server, int) {
	c := config.NewDefaultConfig()
	c.DbPath = fmt.Sprintf("%d.db", time.Now().UnixNano())
	p := pipeline.NewPipeline(c)
	testPort += 1
	s := NewServer(testPort, p, auths)
	return s, testPort
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
