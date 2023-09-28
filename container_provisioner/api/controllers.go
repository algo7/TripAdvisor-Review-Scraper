package api

import (
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/containers"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/database"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/utils"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var R2Url = "https://storage.algo7.tools/"

type EnrichedR2Objs struct {
	FileName   string
	Link       string
	UploadedBy string
	Date       string
}
/
type RunningTask struct {
	ContainerId string
	Url         string
	TaskOwner   string
	HotelName   string
}

// getMain renders the main page
func getMain(c *fiber.Ctx) error {

	// Get the number of running containers
	runningContainers := containers.CountRunningContainer()

	// Check if the R2 objects list is cached
	cachedObjectsList := database.CacheLookUp("r2StorageObjectsList")

	// If the R2 objects list is cached, return the cached value
	if cachedObjectsList != "" {

		// Decode the JSON encoded byte slice into a slice of EnrichedR2Objs structs
		var enrichedR2Objs = []EnrichedR2Objs{}
		err := json.Unmarshal([]byte(cachedObjectsList), &enrichedR2Objs)
		utils.ErrorHandler(err)

		return c.Render("main", fiber.Map{
			"Title":             "Algo7 TripAdvisor Scraper",
			"RunningContainers": runningContainers,
			"Rows":              enrichedR2Objs,
		})

	}

	// If the value is not cached, get the list of objects from R2 and cache it

	// Get the list of objects from the R2 bucket (without metadata)
	r2Objs := utils.R2ListObjects()

	// Enrich the R2 objects with metadata
	R2ObjMetaData := utils.R2EnrichMetaData(r2Objs)

	// Create a slice of Row structs to hold the data for the table
	enrichedR2Objs := make([]EnrichedR2Objs, len(R2ObjMetaData))

	// Populate the slice of Row struct with data from the fileNames array
	for i, r2Obj := range R2ObjMetaData {
		enrichedR2Objs[i] = EnrichedR2Objs{
			FileName:   r2Obj.Key,
			Link:       R2Url + r2Obj.Key,
			UploadedBy: r2Obj.Metadata,
			Date:       utils.ParseTime(r2Obj.LastModified),
		}
	}

	// Store the encoded byte slice into redis
	database.SetCache("r2StorageObjectsList", enrichedR2Objs)

	return c.Render("main", fiber.Map{
		"Title":             "Algo7 TripAdvisor Scraper",
		"RunningContainers": runningContainers,
		"Rows":              enrichedR2Objs,
	})
}

// postProvision is the handler for the form submission
func postProvision(c *fiber.Ctx) error {

	// Get the URL from the form
	url := c.FormValue("url")

	// Get the email from the form
	uploadIdentifier := c.FormValue("upload_identifier")

	// Validate the uploadIdentifier field
	if uploadIdentifier == "" || len(uploadIdentifier) > 20 {
		return c.Render("submission", fiber.Map{
			"Title":      "Algo7 TripAdvisor Scraper",
			"Message1":   "Please provide a valid identifier",
			"ReturnHome": true,
		})
	}

	// Check if the URL matches the regex
	if !utils.ValidateTripAdvisorHotelURL(url) {
		return c.Render("submission", fiber.Map{
			"Title":      "Algo7 TripAdvisor Scraper",
			"Message1":   "Invalid URL",
			"ReturnHome": true,
		})
	}

	// Get the number of running containers
	runningContainers := containers.CountRunningContainer()

	if runningContainers >= 10 {
		return c.Render("submission", fiber.Map{
			"Title":      "Algo7 TripAdvisor Scraper",
			"Message1":   "Sorry, we are currently busy. Please try again later",
			"ReturnHome": true,
		})
	}

	// Get the hotel name from the URL
	hotelName := utils.GetHotelNameFromURL(url)

	// Create the container
	containerId := containers.CreateContainer(hotelName, url, uploadIdentifier)

	// Start the scraping container via goroutine
	go containers.Scrape(uploadIdentifier, hotelName, containerId)

	return c.Render("submission", fiber.Map{
		"Title": "Algo7 TripAdvisor Scraper",
		// "Message": fmt.Sprintf("Your request has been submitted. You will receive an email at %s when the data is ready", email),
		"Message1":    "Your request has been submitted. ",
		"Message2":    "You can check the progress of your request below",
		"Message3":    "Once it's done, you can return to the main page to download the data",
		"ContainerId": containerId,
		"UploadID":    fmt.Sprintf("Your Upload ID: %s", uploadIdentifier),
		"ReturnHome":  false,
		// "URL":      R2Url + fileSuffix + "-" + "0" + "_" + hotelName + ".csv",

	})
}

// getLogsViewer renders the logs viewer page
func getLogsViewer(c *fiber.Ctx) error {
	return c.SendFile("./views/logs.html")
}

// getLogs returns the logs for a given container
func getLogs(c *fiber.Ctx) error {
	containerId := c.Params("id")
	if containerId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Container ID is required"})
	}

	// Get ids of all running containers
	existingContainers := containers.ListContainers()

	// If there are no running containers
	if len(existingContainers) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No containers are running right now"})
	}

	// Existing container ids
	runningContainersIds := []string{}

	// Extract the running container ids
	for _, container := range existingContainers {
		runningContainersIds = append(runningContainersIds, container.ID)
	}

	// If the running containers do not include the containerId
	if !strings.Contains(strings.Join(runningContainersIds, ","), containerId) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Container ID is invalid"})
	}

	// Get the logs for the container
	logsReader := containers.TailLog(containerId)

	// Send the stream to the client
	return c.SendStream(logsReader)
}

// getRunningJobs renders a table of running containers
func getRunningTasks(c *fiber.Ctx) error {

	// Get ids of all running containers
	Containers := containers.ListContainers()

	// Create a slice of RunningTask structs to hold the data for the table
	runningTasks := make([]RunningTask, len(Containers))

	// Populate the slice of RunningTask structs with data from the Containers array
	for i, container := range Containers {
		// Exclude the container that runs app itself
		if container.TaskOwner != "" {
			runningTasks[i] = RunningTask{
				ContainerId: container.ID[:12],
				Url:         fmt.Sprintf("/logs-viewer?container_id=%s", container.ID),
				TaskOwner:   container.TaskOwner,
				HotelName:   container.HotelName,
			}
		}
	}

	// The page status message
	currentTaskStatus := "There are no running tasks"

	if len(Containers) > 0 {
		currentTaskStatus = fmt.Sprintf("%s task(s) running", strconv.Itoa(len(Containers)-2))
	}

	return c.Render("tasks", fiber.Map{
		"Title":             "Algo7 TripAdvisor Scraper",
		"RunningTasks":      runningTasks,
		"CurrentTaskStatus": currentTaskStatus,
	})
}
