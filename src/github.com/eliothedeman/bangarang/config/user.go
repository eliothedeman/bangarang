package config

import (
	"encoding/binary"
)

// UserPermissions describes what a user is allowed to do
type UserPermissions int

const (
	READ = iota
	WRITE
	ADMIN // bless users to higher permissions levels
)

// User holds information about who a user is, and what they are allowed to do
type User struct {
	Id           uint16          `json:"id"`
	Name         string          `json:"name"`
	Email        string          `json:"email"`
	PasswordHash string          `json:"password_hash"`
	Permissions  UserPermissions `json:"permissions"`
}

// encode the id as binary
func idToBin(id uint16) []byte {
	b := []byte{0, 0}
	binary.BigEndian.PutUint16(b, id)
	return b
}

func binToId(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}
