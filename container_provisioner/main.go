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
)

const containerImage = "ghcr.io/algo7/tripadvisor-review-scraper/scraper:latest"

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

	// Pull the scraper image
	err = cm.PullImage()
	if err != nil {
		log.Fatalf("fail to pull the scraper image: %s", err)
	}

	// Set up signal handling to catch SIGINT and SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Launch a goroutine that will perform cleanup when a signal is received
	go func() {
		// sig := <-sigCh
		<-sigCh
		err := cleanupScraperContainers(cm)
		if err != nil {
			log.Printf("cleanup failed: %v", err)
		}
		cm.Close()
		app.Shutdown() // gracefully stops the listener
		// os.Exit(int(sig.(syscall.Signal)))
	}()

	err = app.Listen(":3000")
	if err != nil {
		log.Printf("server stopped: %v", err)
	}
}

// cleanupScraperContainers removes all the running scraper containers
func cleanupScraperContainers(c *containers.ContainerManager) error {

	runningScrapers, err := c.ListContainersByType("scraper")
	if err != nil {
		return fmt.Errorf("fail to list running scraper containers: %w", err)
	}

	for _, container := range runningScrapers {
		err := c.RemoveContainer(*container.ContainerID)
		if err != nil {
			log.Printf("fail to remove container %s: %v", *container.ContainerID, err)
		} else {
			log.Printf("successfully removed container %s", *container.ContainerID)
		}
	}

	return nil
}
