package api

import (
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/scrape"
	"github.com/gofiber/fiber/v2"
)

// Router is the main router for the API
func Router(s *scrape.Scraper) *fiber.App {
	h := &Handler{s}
	app := ServerInstantiate()

	app.Get("/", h.getMain)
	app.Post("/submit", h.postProvision)
	app.Get("/logs/:id", h.getLogs)
	app.Get("/logs-viewer", h.getLogsViewer)
	app.Get("/tasks", h.getRunningTasks)
	app.Get("/downloads", h.getDownloads)

	return app
}
