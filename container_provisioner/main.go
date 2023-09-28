package main

import (
	"log"
	"os"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/api"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/database"
)

func main() {

	// Check if the R2_URL environment variable is set
	if os.Getenv("R2_URL") == "" {
		log.Fatal("R2_URL environment variable not set")
	}

	// Check if the redis server is up and running
	database.RedisConnectionCheck()

	// Load the API routes
	api.Router()

}
