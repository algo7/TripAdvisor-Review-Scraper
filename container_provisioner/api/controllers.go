package api

import (
	"container_provisioner/containers"
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

	// Get the number of running containers
	runningContainers := containers.CountRunningContainer()

	if runningContainers >= 5 {
		return c.SendString("Sorry, we are currently busy. Please try again later")
	}

	// Provision the container
	go containers.Provision(url)

	return c.SendString("Your file will be to your email address when finished 👋")
}
