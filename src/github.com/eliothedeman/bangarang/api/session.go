package api

import (
	"errors"
	"sync"
	"time"

	"github.com/satori/go.uuid"
)

func init() {
	GlobalSession = NewSession()
}

var (
	defaultExparation   = time.Hour * 48
	GlobalSession       *Session
	InvalidSessionToken = errors.New("Invalid Session Token")
	ExpiredSessionToken = errors.New("Session Token Expired")
)

type Session struct {
	sync.RWMutex
	store map[string]SessionToken
}

func NewSession() *Session {
	return &Session{
		store: map[string]SessionToken{},
	}
}

// Get a User.UserName from the session store that corrospondes with the session token
func (s *Session) Get(token string) (string, error) {
	s.Lock()
	defer s.Unlock()
	t, ok := s.store[token]

	// no token found
	if !ok {
		return "", InvalidSessionToken
	}

	// if the token is expired, remove it, and return the error
	if t.expire.Before(time.Now()) {

		// remove the token from the store
		delete(s.store, token)

		return "", ExpiredSessionToken
	}
	return t.userName, nil
}

// Delete the token from the session store
func (s *Session) Delete(token string) {
	s.Lock()
	defer s.Unlock()

	delete(s.store, token)
}

// Put a new session token into the store for the given user id, and return the token
func (s *Session) Put(userName string) string {
	s.Lock()
	defer s.Unlock()
	t := NewSessionToken(userName)

	s.store[t.token] = t
	return t.token
}

// NewSessionToken create a unique session token for the given user id
func NewSessionToken(userName string) SessionToken {

	// create a new uuid
	u := uuid.NewV4()

	// use that uuid as the token
	t := SessionToken{
		token:    u.String(),
		userName: userName,
		expire:   time.Now().Add(defaultExparation),
	}

	return t
}

type SessionToken struct {
	token    string
	userName string
	expire   time.Time
}
