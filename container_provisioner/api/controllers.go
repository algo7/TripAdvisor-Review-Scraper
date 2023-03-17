package api

import (
	"container_provisioner/containers"
	"container_provisioner/utils"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

var R2Url = "https://storage.algo7.tools/"

type Row struct {
	FileName   string
	Link       string
	UploadedBy string
}

// getMain renders the main page
func getMain(c *fiber.Ctx) error {

	// Get the list of objects from the R2 bucket (without metadata)
	r2Objs := utils.R2ListObjects()

	// Enrich the R2 objects with metadata
	enrichedR2Objs := utils.R2EnrichMetaData(r2Objs)

	// Create a slice of Row structs to hold the data for the table
	rows := make([]Row, len(enrichedR2Objs))

	// Populate the rows slice with data from the fileNames array
	for i, r2Obj := range enrichedR2Objs {
		rows[i] = Row{
			FileName:   r2Obj.Key,
			Link:       R2Url + r2Obj.Key,
			UploadedBy: r2Obj.Metadata,
		}
	}

	// Get the number of running containers
	runningContainers := containers.CountRunningContainer()

	return c.Render("main", fiber.Map{
		"Title":             "Algo7 TripAdvisor Scraper",
		"RunningContainers": runningContainers,
		"Rows":              rows,
	})
}

// postProvision is the handler for the form submission
func postProvision(c *fiber.Ctx) error {

	// Get the URL from the form
	url := c.FormValue("url")

	// Get the email from the form
	uploadIdentifier := c.FormValue("upload_identifier")

	// Validate the uploadIdentifier field
	if uploadIdentifier == "" {
		return c.Render("submission", fiber.Map{
			"Title":    "Algo7 TripAdvisor Scraper",
			"Message1": "Please provide the identifier for the data",
		})
	}

	// Check if the URL matches the regex
	if !utils.ValidateTripAdvisorHotelURL(url) {
		return c.Render("submission", fiber.Map{
			"Title":    "Algo7 TripAdvisor Scraper",
			"Message1": "Invalid URL",
		})
	}

	// Get the number of running containers
	runningContainers := containers.CountRunningContainer()

	if runningContainers >= 10 {
		return c.Render("submission", fiber.Map{
			"Title":    "Algo7 TripAdvisor Scraper",
			"Message1": "Sorry, we are currently busy. Please try again later",
		})
	}

	// Generate a random file prefix
	fileSuffix := utils.GenerateUUID()

	// Get the hotel name from the URL
	// hotelName := utils.GetHotelNameFromURL(url)

	// Provision the container via goroutine
	go containers.Provision(fileSuffix, uploadIdentifier, url)

	return c.Render("submission", fiber.Map{
		"Title": "Algo7 TripAdvisor Scraper",
		// "Message": fmt.Sprintf("Your request has been submitted. You will receive an email at %s when the data is ready", email),
		"Message1": "Your request has been submitted. ",
		"Message2": "Return to the Home Page and Check for Your File.",
		"Message3": "Your data should be available shortly.",
		"UploadID": fmt.Sprintf("Your Upload ID: %s", uploadIdentifier),
		// "URL":      R2Url + fileSuffix + "-" + "0" + "_" + hotelName + ".csv",
	})
}
