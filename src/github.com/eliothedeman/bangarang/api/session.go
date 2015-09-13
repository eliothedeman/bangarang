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
	defaultExparation = time.Hour * 48
	GlobalSession     *Session
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

// Get a User.Id from the session store that corrospondes with the session token
func (s *Session) Get(token string) (uint16, error) {
	s.Lock()
	defer s.Unlock()
	t, ok := s.store[token]

	// no token found
	if !ok {
		return 0, errors.New("Invalid session token")
	}

	// if the token is expired, remove it, and return the error
	if t.expire.Before(time.Now()) {

		// remove the token from the store
		delete(s.store, token)

		return 0, errors.New("Session token expired")
	}
	return t.userId, nil
}

// Delete the token from the session store
func (s *Session) Delete(token string) {
	s.Lock()
	defer s.Unlock()

	delete(s.store, token)
}

// Put a new session token into the store for the given user id, and return the token
func (s *Session) Put(id uint16) string {
	s.Lock()
	defer s.Unlock()
	t := NewSessionToken(id)

	s.store[t.token] = t
	return t.token
}

// NewSessionToken create a unique session token for the given user id
func NewSessionToken(id uint16) SessionToken {

	// create a new uuid
	u := uuid.NewV4()

	// use that uuid as the token
	t := SessionToken{
		token:  u.String(),
		userId: id,
		expire: time.Now().Add(defaultExparation),
	}

	return t
}

type SessionToken struct {
	token  string
	userId uint16
	expire time.Time
}
