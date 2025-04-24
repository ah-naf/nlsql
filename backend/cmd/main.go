package main

import (
	"nlsql/config"
	"nlsql/internal/api"
)

func main() {
	// Load .env
	config.LoadEnv()

	// Start HTTP server
	router := api.SetupRouter()
	router.Run() // defaults to :8080
}
