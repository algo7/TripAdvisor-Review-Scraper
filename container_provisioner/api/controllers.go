package api

import (
	"container_provisioner/utils"

	"github.com/gofiber/fiber/v2"
)

// getMain renders the main page
func getMain(c *fiber.Ctx) error {
	return c.Render("main", fiber.Map{})
}

// postProvision is the handler for the form submission
func postProvision(c *fiber.Ctx) error {

	// Get the URL from the form
	url := c.FormValue("url")

	// Check if the URL matches the regex
	if !utils.ValidateTripAdvisorHotelURL(url) {
		return c.SendString("Invalid URL")
	}

	return c.SendString("Hello, World 👋!" + url)
}
