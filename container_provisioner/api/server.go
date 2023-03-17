package api

import (
	"github.com/gofiber/fiber/v2"
)

var (
	// Custom config
	App = fiber.New(fiber.Config{
		Prefork:       true,
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Algo7 TripAdvisor Scraper",
		AppName:       "Algo7 TripAdvisor Scraper v1.0.0",
	})
)

func ServerInit() {
	App.Listen(":3000")
}
