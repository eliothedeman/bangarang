package config

import "testing"

func TestNewUser(t *testing.T) {
	u := NewUser("Test User", "albert", "password", ADMIN)

	if u.Name != "Test User" {
		t.Fatal()
	}

	if u.UserName != "albert" {
		t.Fatal()
	}

	if u.Permissions != ADMIN {
		t.Fatal()
	}
}

func TestPasswordCheck(t *testing.T) {
	u := NewUser("Test User", "albert", "password", ADMIN)

	if !CheckUserPassword(u, "password") {
		t.Fatal()
	}
}
