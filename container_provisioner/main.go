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
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/scrape"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/storage"
	"github.com/gofiber/fiber/v2"
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
		log.Fatalf("fail to initialize container manager: %s", err)
	}

	// Initialize the storage client
	r2, err := storage.NewR2Service("./credentials/creds.json")
	if err != nil {
		log.Fatalf("fail to initialize R2 service: %v", err)
	}

	// Initialize the scraper
	scraper := scrape.NewScraper(cm, r2, r)

	// Load the API routes
	app := api.Router(scraper)

	if !fiber.IsChild() {
		// Try to acquire the lock for pullig container image
		lockSuccess := r.SetLock(imageLockKey)
		if lockSuccess {
			err = cm.PullImage()
			r.ReleaseLock(imageLockKey)
			if err != nil {
				log.Fatalf("fail to pull the scraper image: %s", err)
			}
		} else {
			log.Println("image pull lock not acquired, skipping pull")
		}

		// Set up signal handling to catch SIGINT and SIGTERM
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		// Launch a goroutine that will perform cleanup when a signal is received
		go func() {
			// sig := <-sigCh
			<-sigCh
			err := cleanupScraperContainers(r, cm)
			if err != nil {
				log.Printf("cleanup failed: %v", err)
			}
			cm.Close()
			app.Shutdown() // gracefully stops the listener
			// os.Exit(int(sig.(syscall.Signal)))
		}()
	}

	err = app.Listen(":3000")
	if err != nil {
		log.Printf("server stopped: %v", err)
	}
}

// cleanupScraperContainers removes all the running scraper containers
func cleanupScraperContainers(r *database.RedisClient, c *containers.ContainerManager) error {

	runningScrapers, err := c.ListContainersByType("scraper")
	if err != nil {
		return fmt.Errorf("fail to list running scraper containers: %w", err)
	}

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

	return nil
}
