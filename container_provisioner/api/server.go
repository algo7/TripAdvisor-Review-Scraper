package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
)

func ServerInstantiate() *fiber.App {

	// Initialize standard Go html template engine
	engine := html.New("./views", ".html")

	// Layout defines the variable name that is used to yield templates within layouts
	engine.Layout("embed") // Optional. Default: "embed"

	// Delims sets the action delimiters to the specified strings
	engine.Delims("{{", "}}") // Optional. Default: engine delimiters

	// Custom config
	app := fiber.New(fiber.Config{
		Prefork:       true,
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Algo7 TripAdvisor Scraper",
		AppName:       "Algo7 TripAdvisor Scraper v1.0.0",
		Views:         engine,
	})

	return app
}
