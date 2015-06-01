package config

import "log"

func DryRun(a *AppConfig) error {
	log.Println(a)
	return nil
}
