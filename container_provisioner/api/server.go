package api

import (
	"container_provisioner/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
)

var (

	// Initialize standard Go html template engine
	engine = html.New("./views", ".html")

	// Custom config
	App = fiber.New(fiber.Config{
		Prefork:       true,
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Algo7 TripAdvisor Scraper",
		AppName:       "Algo7 TripAdvisor Scraper v1.0.0",
		Views:         engine,
	})
)

func ServerInit() {
	err := App.Listen(":3000")
	utils.ErrorHandler(err)
}
