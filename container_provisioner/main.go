package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/api"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/containers"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/database"
)

const containerImage = "ghcr.io/algo7/tripadvisor-review-scraper/scraper:latest"

var imageLockKey = fmt.Sprintf("image-pull:%s", containerImage)

func main() {

	// Check if the R2_URL environment variable is set
	if os.Getenv("R2_URL") == "" {
		log.Fatal("R2_URL environment variable not set")
	}

	// Initialize the Redis client
	r := database.NewRedisClient()

	resp, err := r.CheckConnection()
	if err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}
	fmt.Println("Redis connection established:", resp)

	//  Initialize container manager
	cm, err := containers.NewContainerManager(containerImage)
	if err != nil {
		log.Fatalf("fail to initialize container manager: %w", err)
	}

	// Try to acquire the lock for pullig container image
	lockSuccess := r.SetLock(imageLockKey)
	if !lockSuccess {
		// If the lock is not acquired, another instance is already pulling the image
		return
	}

	// Pull the scraper image
	err = cm.PullImage()
	if err != nil {
		log.Fatalf("fail to pull the scraper image: %w", err)
	}

	// Set up signal handling to catch SIGINT and SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Launch a goroutine that will perform cleanup when a signal is received
	go func() {
		sig := <-sigCh
		cleanupScraperContainers(r, cm)
		err := r.ReleaseLock(imageLockKey)
		if err != nil {
			log.Printf("fail to release lock after cleanup for image %s: %v", containerImage, err)
		} else {
			log.Printf("successfully released lock after cleanup for image %s", containerImage)
		}
		os.Exit(int(sig.(syscall.Signal)))
	}()

	// Load the API routes
	api.Router()
}

// cleanupScraperContainers removes all the running scraper containers
func cleanupScraperContainers(r *database.RedisClient, c *containers.ContainerManager) {

	runningScrapers := containers.ListContainersByType("scraper")

	for _, container := range runningScrapers {

		lockKey := "container-cleanup:" + *container.ContainerID
		lockSuccess := r.SetLock(lockKey)
		if !lockSuccess {
			continue // skip to the next iteration of the loop
		}
		// If lockSuccess is true, we have the lock, so we can proceed with the cleanup
		err := c.RemoveContainer(*container.ContainerID)
		if err != nil {
			log.Printf("fail to remove container %s: %v", *container.ContainerID, err)
		} else {
			log.Printf("successfully removed container %s", *container.ContainerID)
		}

		// Release the lock after cleanup
		err = r.ReleaseLock(lockKey)
		if err != nil {
			log.Printf("fail to release lock for container %s: %v", *container.ContainerID, err)
		} else {
			log.Printf("successfully released lock for container %s", *container.ContainerID)
		}

	}
}
