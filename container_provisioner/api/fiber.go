package api

import (
	"github.com/gofiber/fiber/v2"
)

// Custom config
app := fiber.New(fiber.Config{
    Prefork:       true,
    CaseSensitive: true,
    StrictRouting: true,
    ServerHeader:  "Algo7 TripAdvisor Scraper",
    AppName: "Algo7 TripAdvisor Scraper v1.0.0"
})