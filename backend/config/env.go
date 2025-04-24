package config

import (
	"log"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file (if present).
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: could not load .env file: %v", err)
	}
}
