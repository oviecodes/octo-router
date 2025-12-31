package main

import (
	"llm-router/cmd/internal/metrics"
	"llm-router/cmd/internal/server"
	"log"
	"path/filepath"

	"github.com/joho/godotenv"
)

func main() {
	envPaths := []string{
		".env",
		"../../.env",
		"../.env",
		filepath.Join(".", ".env"),
	}

	loaded := false
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			log.Printf("Loaded environment variables from %s", path)
			loaded = true
			break
		}
	}

	if !loaded {
		log.Println("No .env file found, using existing environment variables")
	}

	server.Server()

	// for now start metrics server immediately
	// later will have the user config enable / disable metrics endpoints
	metrics.Metrics()
}
