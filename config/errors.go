package config

import "fmt"

type ConfigVersionNotFound string
type UserNotFound string

func (c ConfigVersionNotFound) Error() string {
	return fmt.Sprintf("Config version %s not found", c)
}

func (u UserNotFound) Error() string {
	return fmt.Sprintf("Config version %s not found", u)
}
