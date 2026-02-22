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
		"ALLOWED_ORIGINS",
		"SPOTIFY_CLIENT_ID",
		"SPOTIFY_CLIENT_SECRET",
		"SPOTIFY_REFRESH_TOKEN",
	}

	for _, env := range(envs) {
		if os.Getenv(env) == "" {
			log.Fatal("Missing required:", env)
		}
	}
}