package util

import (
	"log"
	"os"
)

func RequireEnvs () {
	envs := []string{
		"PORT",
		"DATABASE_URL",
		"RESEND_API_KEY",
		"ALLOWED_ORIGIN",
	}

	for _, env := range(envs) {
		if os.Getenv(env) == "" {
			log.Fatal("Missing required:", env)
		}
	}
}