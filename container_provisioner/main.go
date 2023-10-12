package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/api"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/containers"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/database"
)

func main() {

	// Set up signal handling to catch SIGINT and SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Launch a goroutine that will perform cleanup when a signal is received
	go func() {
		sig := <-sigCh
		cleanupScraperContainers()
		os.Exit(int(sig.(syscall.Signal)))
	}()

	// Check if the R2_URL environment variable is set
	if os.Getenv("R2_URL") == "" {
		log.Fatal("R2_URL environment variable not set")
	}

	// Check if the redis server is up and running
	database.RedisConnectionCheck()

	// Load the API routes
	api.Router()

}

// cleanupScraperContainers removes all the running scraper containers
func cleanupScraperContainers() {

	runningScrapers := containers.ListContainersByType("scraper")

	for _, container := range runningScrapers {

		lockKey := "container-cleanup:" + *container.ContainerID
		lockSuccess := database.SetLock(lockKey)
		if !lockSuccess {
			continue // skip to the next iteration of the loop
		}

		// If lockSuccess is true, we have the lock, so we can proceed with the cleanup
		containers.RemoveContainer(*container.ContainerID)
		database.ReleaseLock(lockKey)

	}
}
