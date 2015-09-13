package config

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
)

// UserPermissions describes what a user is allowed to do
type UserPermissions int

func InsufficientPermissions(need, have UserPermissions) error {
	return fmt.Errorf("Insufficient permissions: You need %s, but have %s", PermissionsToName(need), PermissionsToName(have))
}

func PermissionsToName(p UserPermissions) string {
	switch p {
	case READ:
		return "read"
	case WRITE:
		return "write"
	case ADMIN:
		return "admin"
	default:
		return "unknown"
	}

}

const (
	READ = iota
	WRITE
	ADMIN // bless users to higher permissions levels
)

// User holds information about who a user is, and what they are allowed to do
type User struct {
	Id           uint16          `json:"id"`
	Name         string          `json:"name"`
	UserName     string          `json:"user_name"`
	PasswordHash string          `json:"password_hash"`
	Permissions  UserPermissions `json:"permissions"`
	provider     *Provider
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

// CheckUserPassword compares a raw password against the the stored hash'ed password
func CheckUserPassword(u *User, pass string) bool {
	return HashUserPassword(u, pass) == u.PasswordHash
}

func NewUser(id uint16, name, userName, rawPassword string, permissions UserPermissions) *User {
	u := &User{
		Id:          id,
		Name:        name,
		UserName:    userName,
		Permissions: permissions,
	}

	u.PasswordHash = HashUserPassword(u, rawPassword)
	return u
}

// return the password hash for a given user, given the raw password
func HashUserPassword(u *User, raw string) string {
	m := md5.New()
	m.Write([]byte(u.UserName + raw))
	return fmt.Sprintf("%x", m.Sum(nil))

}
