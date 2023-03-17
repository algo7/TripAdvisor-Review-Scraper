package api

import "github.com/gofiber/fiber/v2"

// mainView renders the main page
func mainView(c *fiber.Ctx) error {
	return c.Render("main", fiber.Map{
		"Title": "Algo7 TripAdvisor Scraper",
	})
}
