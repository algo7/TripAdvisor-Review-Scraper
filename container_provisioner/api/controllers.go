package api

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/containers"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/database"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/utils"

	"github.com/gofiber/fiber/v2"
)

// The URL of the R2 bucket
var r2Url string = os.Getenv("R2_URL")

type enrichedR2Obj struct {
	FileName   string
	Link       string
	UploadedBy string
	Date       string
}

// getMain renders the main page
func getMain(c *fiber.Ctx) error {

	// Get the number of running containers
	runningContainers := len(containers.ListContainersByType("scraper"))

	return c.Render("main", fiber.Map{
		"Title":             "Algo7 TripAdvisor Scraper",
		"RunningContainers": runningContainers,
	})
}

// postProvision is the handler for the form submission
func postProvision(c *fiber.Ctx) error {

	// Get the URL from the form
	url := c.FormValue("url")

	// Get the upload id from the form
	uploadIdentifier := c.FormValue("upload_identifier")

	// Get the scrape mode from the form
	scrapeMode := c.FormValue("scrape_option")

	// Define valid scrape modes
	validChoices := map[string]bool{
		"HOTEL":   true,
		"RESTO":   true,
		"AIRLINE": true,
	}

	_, exists := validChoices[scrapeMode]

	// Validate the scrape mode
	if !exists {
		return c.Render("submission", fiber.Map{
			"Title":      "Algo7 TripAdvisor Scraper",
			"Message1":   "Invalid Scrape Target",
			"ReturnHome": true,
		})
	}

	// Check if the URL is valid
	if !utils.ValidateTripAdvisorURL(url, scrapeMode) {
		return c.Render("submission", fiber.Map{
			"Title":      "Algo7 TripAdvisor Scraper",
			"Message1":   fmt.Sprintf("Invalid %s URL", scrapeMode),
			"ReturnHome": true,
		})
	}

	// Validate the uploadIdentifier field
	if uploadIdentifier == "" || len(uploadIdentifier) > 20 {
		return c.Render("submission", fiber.Map{
			"Title":      "Algo7 TripAdvisor Scraper",
			"Message1":   "Please provide a valid identifier",
			"ReturnHome": true,
		})
	}

	// Get the number of running containers
	runningContainers := len(containers.ListContainersByType("scraper"))

	if runningContainers >= 5 {
		return c.Render("submission", fiber.Map{
			"Title":      "Algo7 TripAdvisor Scraper",
			"Message1":   "Sorry, we are currently busy. Please try again later",
			"ReturnHome": true,
		})
	}

	// Get the location name
	locationName := utils.GetLocationNameFromURL(url, scrapeMode)
	if locationName == "" {
		return c.Render("submission", fiber.Map{
			"Title":      "Algo7 TripAdvisor Scraper",
			"Message1":   "Invalid URL",
			"ReturnHome": true,
		})
	}

	// Get the proxy container info
	proxyContainers := containers.AcquireProxyContainer()

	// Generate the container config
	scrapeConfig := containers.ContainerConfigGenerator(
		url,
		locationName,
		uploadIdentifier,
		proxyContainers.ProxyAddress,
		proxyContainers.VPNRegion)

	// Create the container
	containerID := containers.CreateContainer(scrapeConfig)

	// Start the scraping container via goroutine
	go func() {
		containers.Scrape(uploadIdentifier, locationName, containerID)
		containers.ReleaseProxyContainer(proxyContainers.ContainerID)
	}()

	return c.Render("submission", fiber.Map{
		"Title": "Algo7 TripAdvisor Scraper",
		// "Message": fmt.Sprintf("Your request has been submitted. You will receive an email at %s when the data is ready", email),
		"Message1":    fmt.Sprintf("Your request has been submitted. VPN Region: %s", proxyContainers.VPNRegion),
		"Message2":    "You can check the progress of your request below",
		"Message3":    "Once it's done, you can return to the main page to download the data",
		"ContainerId": containerID,
		"UploadID":    fmt.Sprintf("Your Upload ID: %s", uploadIdentifier),
		"ReturnHome":  false,
		// "URL":      r2Url + fileSuffix + "-" + "0" + "_" + hotelName + ".csv",

	})
}

// getLogsViewer renders the logs viewer page
func getLogsViewer(c *fiber.Ctx) error {
	return c.SendFile("./views/logs.html")
}

// getLogs returns the logs for a given container
func getLogs(c *fiber.Ctx) error {
	containerID := c.Params("id")
	if containerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Container ID is required"})
	}

	// Get ids of all running containers
	existingContainers := containers.ListContainersByType("scraper")

	// If there are no running containers
	if len(existingContainers) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No containers are running right now"})
	}

	// Existing container ids
	runningContainersIds := []string{}

	// Extract the running container ids
	for _, container := range existingContainers {
		runningContainersIds = append(runningContainersIds, *container.ContainerID)
	}

	// If the running containers do not include the containerId
	if !strings.Contains(strings.Join(runningContainersIds, ","), containerID) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Container ID is invalid"})
	}

	// Get the logs for the container
	logsReader := containers.TailLog(containerID)

	// Send the stream to the client
	return c.SendStream(logsReader)
}

// getRunningJobs renders a table of running containers
func getRunningTasks(c *fiber.Ctx) error {

	// Get ids of all running containers
	runningContainers := containers.ListContainersByType("scraper")

	// The page status message
	currentTaskStatus := "There are no running tasks"

	if len(runningContainers) != 0 {
		currentTaskStatus = fmt.Sprintf("%s task(s) running", strconv.Itoa(len(runningContainers)))
	}

	return c.Render("tasks", fiber.Map{
		"Title":             "Algo7 TripAdvisor Scraper",
		"RunningTasks":      runningContainers,
		"CurrentTaskStatus": currentTaskStatus,
	})
}

// getDownloads renders the downloads page
func getDownloads(c *fiber.Ctx) error {

	// Check if the R2 objects list is cached
	cachedObjectsList := database.CacheLookUp("r2StorageObjectsList")

	// If the R2 objects list is cached, return the cached value
	if cachedObjectsList != "" {

		// Decode the JSON encoded byte slice into a slice of EnrichedR2Objs structs
		var enrichedR2Objs = []enrichedR2Obj{}
		err := json.Unmarshal([]byte(cachedObjectsList), &enrichedR2Objs)
		utils.ErrorHandler(err)

		return c.Render("downloads", fiber.Map{
			"Title": "Algo7 TripAdvisor Scraper",
			"Rows":  enrichedR2Objs,
		})

	}

	// If the value is not cached, get the list of objects from R2 and cache it

	// Get the list of objects from the R2 bucket (without metadata)
	r2Objs := utils.R2ListObjects()

	// Enrich the R2 objects with metadata
	R2ObjMetaData := utils.R2EnrichMetaData(r2Objs)

	// Create a slice of Row structs to hold the data for the table
	enrichedR2Objs := make([]enrichedR2Obj, len(R2ObjMetaData))

	// Populate the slice of Row struct with data from the fileNames array
	for i, r2Obj := range R2ObjMetaData {
		enrichedR2Objs[i] = enrichedR2Obj{
			FileName:   r2Obj.Key,
			Link:       r2Url + r2Obj.Key,
			UploadedBy: r2Obj.Metadata,
			Date:       utils.ParseTime(r2Obj.LastModified),
		}
	}

	// Store the encoded byte slice into redis
	database.SetCache("r2StorageObjectsList", enrichedR2Objs)

	return c.Render("main", fiber.Map{
		"Title": "Algo7 TripAdvisor Scraper",
		"Rows":  enrichedR2Objs,
	})
}
