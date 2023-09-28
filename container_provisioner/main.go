package main

import (
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/api"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/database"
)

func main() {

	// Check if the redis server is up and running
	database.RedisConnectionCheck()

	// Load the API routes
	api.Router()

}
