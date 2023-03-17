package api

import (
	"container_provisioner/containers"
	"container_provisioner/utils"

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

	// Get the list of objects from the R2 bucket
	r2Objs := utils.R2ListObjects()

	// Create a slice of Row structs to hold the data for the table
	rows := make([]Row, len(r2Objs))

	// Populate the rows slice with data from the fileNames array
	for i, r2Obj := range r2Objs {
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
			"Title":   "Algo7 TripAdvisor Scraper",
			"Message": "Please provide the identifier for the data",
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

	// Generate a random file prefix
	filePrefix := utils.GenerateUUID()

	// Get the hotel name from the URL
	// hotelName := utils.GetHotelNameFromURL(url)

	// Provision the container via goroutine
	go containers.Provision(filePrefix, uploadIdentifier, url)

	return c.Render("submission", fiber.Map{
		"Title": "Algo7 TripAdvisor Scraper",
		// "Message": fmt.Sprintf("Your request has been submitted. You will receive an email at %s when the data is ready", email),
		"Message1": "Your request has been submitted. ",
		"Message2": "Return to the Home Page and Check for Your File.",
		"Message3": "Your data should be available shortly.",
		"UploadID": uploadIdentifier,
		// "URL":      R2Url + filePrefix + "-" + "0" + "_" + hotelName + ".csv",
	})
}
