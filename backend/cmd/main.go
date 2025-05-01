package main

import (
	"log"
	"nlsql/config"
	"nlsql/internal/api"
	"nlsql/internal/db"
	"os"
	"time"
)

func main() {
	// Load .env
	config.LoadEnv()

	// Start HTTP server
	router := api.SetupRouter()

	db.SetupDemoDBCleanup(1*time.Hour, 2*time.Hour)

	if _, err := os.Stat("../frontend/dist"); os.IsNotExist(err) {
		log.Println("Warning: Frontend build directory not found. Make sure to build the React app first.")
		log.Println("Run: cd frontend && npm run build")
	}

	log.Println("Server starting on :8080...")
	router.Run() // defaults to :8080
}
