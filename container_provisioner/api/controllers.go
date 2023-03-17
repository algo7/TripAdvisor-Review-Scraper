package api

import (
	"container_provisioner/containers"
	"container_provisioner/utils"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// getMain renders the main page
func getMain(c *fiber.Ctx) error {

	// Get the number of running containers
	runningContainers := containers.CountRunningContainer()

	return c.Render("main", fiber.Map{
		"Title":             "Algo7 TripAdvisor Scraper",
		"RunningContainers": runningContainers,
	})
}

// postProvision is the handler for the form submission
func postProvision(c *fiber.Ctx) error {

	// Get the URL from the form
	url := c.FormValue("url")

	// Get the email from the form
	email := c.FormValue("email")

	utils.R2ListObjects()

	// Validate the email
	if !utils.ValidateEmailAddress(email) {
		return c.Render("submission", fiber.Map{
			"Title":   "Algo7 TripAdvisor Scraper",
			"Message": "Invalid email address",
		})
	}

	// Check if the URL matches the regex
	if !utils.ValidateTripAdvisorHotelURL(url) {
		return c.Render("submission", fiber.Map{
			"Title":   "Algo7 TripAdvisor Scraper",
			"Message": "Invalid URL",
		})
	}

	// Get the number of running containers
	runningContainers := containers.CountRunningContainer()

	if runningContainers >= 5 {
		return c.Render("submission", fiber.Map{
			"Title":   "Algo7 TripAdvisor Scraper",
			"Message": "Sorry, we are currently busy. Please try again later",
		})
	}
	// Generate a random file name
	filePrefix := utils.GenerateUUID()

	// Provision the container via goroutine
	go containers.Provision(filePrefix, url)

	return c.Render("submission", fiber.Map{
		"Title":   "Algo7 TripAdvisor Scraper",
		"Message": fmt.Sprintf("Your request has been submitted. You will receive an email at %s when the data is ready", email),
	})
}
