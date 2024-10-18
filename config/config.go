package config

import (
	"os"
)

const PORT = "8080"

func GetPort() string {
	if port := os.Getenv("PORT"); port == "" {
		port = PORT
	}
	return ":" + PORT
}
