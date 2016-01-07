package client

import "testing"

func TestGenUrl(t *testing.T) {
	c := NewClient()

	url := c.GenUrl("api/auth/user", map[string]string{"hello": "world"})

	if url != "http://localhost:8081/api/auth/user?hello=world" {
		t.Fail()
	}
}
