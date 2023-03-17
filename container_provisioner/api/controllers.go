package api

import "github.com/gofiber/fiber/v2"

// getMain renders the main page
func getMain(c *fiber.Ctx) error {
	return c.Render("main", fiber.Map{})
}

// postProvision is the handler for the form submission
func postProvision(c *fiber.Ctx) error {

	url := c.FormValue("url")
	return c.SendString("Hello, World 👋!" + url)
}
