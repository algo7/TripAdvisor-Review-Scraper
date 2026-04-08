package api

import (
	"log"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/scrape"
)

// Router is the main router for the API
func Router(scraper *scrape.Scraper) {

	app := ServerInstantiate()

	app.Get("/", getMain)
	app.Post("/submit", postProvision)
	app.Get("/logs/:id", getLogs)
	app.Get("/logs-viewer", getLogsViewer)
	app.Get("/tasks", getRunningTasks)
	app.Get("/downloads", getDownloads)

	err := app.Listen(":3000")
	if err != nil {
		log.Fatalf("unable to start the router: %v", err)
	}

}
