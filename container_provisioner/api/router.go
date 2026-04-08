package api

import (
	"log"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/scrape"
)

// Router is the main router for the API
func Router(s *scrape.Scraper) {
	h := &Handler{s}
	app := ServerInstantiate()

	app.Get("/", h.getMain)
	app.Post("/submit", h.postProvision)
	app.Get("/logs/:id", h.getLogs)
	app.Get("/logs-viewer", h.getLogsViewer)
	app.Get("/tasks", h.getRunningTasks)
	app.Get("/downloads", h.getDownloads)

	err := app.Listen(":3000")
	if err != nil {
		log.Fatalf("unable to start the router: %v", err)
	}

}
