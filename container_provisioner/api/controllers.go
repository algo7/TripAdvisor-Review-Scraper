package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/scrape"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/utils"

	"github.com/gofiber/fiber/v3"
)

// The URL of the R2 bucket
var r2Url string = os.Getenv("R2_URL")

type enrichedR2Obj struct {
	FileName   string
	Link       string
	UploadedBy string
	Date       string
}

type Handler struct {
	*scrape.Scraper
}

// getMain renders the main page
func (h *Handler) getMain(c fiber.Ctx) error {
	runningContainers, _ := h.Scraper.CM.ListContainersByType("scraper")
	return c.Render("main", fiber.Map{
		"Title":             appTitle,
		"RunningContainers": len(runningContainers),
	})
}

// postProvision is the handler for the form submission
func (h *Handler) postProvision(c fiber.Ctx) error {

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
			"Title":      appTitle,
			"Message1":   "Invalid Scrape Target",
			"ReturnHome": true,
		})
	}

	// Check if the URL is valid
	if !utils.ValidateTripAdvisorURL(url, scrapeMode) {
		return c.Render("submission", fiber.Map{
			"Title":      appTitle,
			"Message1":   fmt.Sprintf("Invalid %s URL", scrapeMode),
			"ReturnHome": true,
		})
	}

	// Validate the uploadIdentifier field
	if uploadIdentifier == "" || len(uploadIdentifier) > 20 {
		return c.Render("submission", fiber.Map{
			"Title":      appTitle,
			"Message1":   "Please provide a valid identifier",
			"ReturnHome": true,
		})
	}

	// Get the number of running containers
	scraperContainers, err := h.Scraper.CM.ListContainersByType("scraper")
	if err != nil {
		return c.Render("submission", fiber.Map{
			"Title":      appTitle,
			"Message1":   "Error checking running containers",
			"ReturnHome": true,
		})
	}

	runningContainers := len(scraperContainers)

	if runningContainers >= 5 {
		return c.Render("submission", fiber.Map{
			"Title":      appTitle,
			"Message1":   "Sorry, we are currently busy. Please try again later",
			"ReturnHome": true,
		})
	}

	// Get the location name
	locationName := utils.GetLocationNameFromURL(url, scrapeMode)
	if locationName == "" {
		return c.Render("submission", fiber.Map{
			"Title":      appTitle,
			"Message1":   "Invalid URL",
			"ReturnHome": true,
		})
	}

	// Get the proxy container info
	proxyContainers, err := h.Scraper.AcquireProxyContainer()
	if err != nil {
		return c.Render("submission", fiber.Map{
			"Title":      appTitle,
			"Message1":   "Sorry, we are currently busy. Please try again later",
			"ReturnHome": true,
		})
	}

	// Generate the container config
	scrapeConfig := h.Scraper.CM.ContainerConfigGenerator(
		url,
		locationName,
		uploadIdentifier,
		proxyContainers.ProxyAddress,
		proxyContainers.VPNRegion)

	// Create the container
	containerID, err := h.Scraper.CM.CreateContainer(scrapeConfig)
	if err != nil {
		// No scrape will run, so free the proxy container for the next request
		h.Scraper.ReleaseProxyContainer(proxyContainers)

		return c.Render("submission", fiber.Map{
			"Title":      appTitle,
			"Message1":   "Error creating scrape task",
			"ReturnHome": true,
		})
	}

	// Start the scraping container via goroutine
	go func() {
		// Keep the proxy container locked for as long as the scrape runs
		ctx, stopHolding := context.WithCancel(context.Background())
		go h.Scraper.HoldProxyContainer(ctx, proxyContainers)

		// Nobody else receives the error: the request that started the scrape finished long ago
		err := h.Scraper.Scrape(uploadIdentifier, locationName, containerID)
		if err != nil {
			log.Printf("scrape %s (container %s) failed: %v", uploadIdentifier, containerID, err)
		}

		stopHolding()
		h.Scraper.ReleaseProxyContainer(proxyContainers)
	}()

	return c.Render("submission", fiber.Map{
		"Title": appTitle,
		// "Message": fmt.Sprintf("Your request has been submitted. You will receive an email at %s when the data is ready", email),
		"Message1":    fmt.Sprintf("Your scrape is running through a proxy in %s.", proxyContainers.VPNRegion),
		"Message2":    "You can follow it while it runs:",
		"Message3":    "When it's done, download the file from the Downloads page.",
		"ContainerId": containerID,
		"UploadID":    fmt.Sprintf("Your Upload ID: %s", uploadIdentifier),
		"ReturnHome":  false,
		// "URL":      r2Url + fileSuffix + "-" + "0" + "_" + hotelName + ".csv",

	})
}

// getLogsViewer renders the logs viewer page
func (h *Handler) getLogsViewer(c fiber.Ctx) error {
	return c.Render("logs", fiber.Map{
		"Title": appTitle,
	})
}

// getLogs returns the logs for a given container
func (h *Handler) getLogs(c fiber.Ctx) error {
	containerID := c.Params("id")
	if containerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No scrape selected. Open a scrape from the Running page."})
	}

	// Get ids of all running containers
	existingContainers, err := h.Scraper.CM.ListContainersByType("scraper")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not check the running scrapes. Please reload the page."})
	}

	// If there are no running containers
	if len(existingContainers) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "This scrape isn't running any more. If it finished, its file is on the Downloads page."})
	}

	// Existing container ids
	runningContainersIds := []string{}

	// Extract the running container ids
	for _, container := range existingContainers {
		runningContainersIds = append(runningContainersIds, *container.ContainerID)
	}

	// If the running containers do not include the containerId
	if !strings.Contains(strings.Join(runningContainersIds, ","), containerID) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "This scrape isn't running any more. If it finished, its file is on the Downloads page."})
	}

	// Get the logs for the container
	logsReader, err := h.Scraper.CM.TailLog(containerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not load the log. Please reload the page."})
	}

	// Send the stream to the client
	return c.SendStream(logsReader)
}

// getRunningJobs renders a table of running containers
func (h *Handler) getRunningTasks(c fiber.Ctx) error {

	// Get ids of all running containers
	runningContainers, err := h.Scraper.CM.ListContainersByType("scraper")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Could not check the running scrapes. Please try again in a moment.")
	}

	// The page status message
	currentTaskStatus := "No scrapes are running right now."

	if len(runningContainers) != 0 {
		currentTaskStatus = fmt.Sprintf("%d of 5 scrapes running.", len(runningContainers))
	}

	return c.Render("tasks", fiber.Map{
		"Title":             appTitle,
		"RunningTasks":      runningContainers,
		"CurrentTaskStatus": currentTaskStatus,
	})
}

// getDownloads renders the downloads page
func (h *Handler) getDownloads(c fiber.Ctx) error {

	// Check if the R2 objects list is cached
	cachedObjectsList, err := h.Scraper.Redis.CacheLookUp("r2StorageObjectsList")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Could not load the list of files. Please try again in a moment.")
	}

	// If the R2 objects list is cached, return the cached value
	if cachedObjectsList != "" {

		// Decode the JSON encoded byte slice into a slice of EnrichedR2Objs structs
		var enrichedR2Objs = []enrichedR2Obj{}
		err := json.Unmarshal([]byte(cachedObjectsList), &enrichedR2Objs)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Could not load the list of files. Please try again in a moment.")
		}

		return c.Render("downloads", fiber.Map{
			"Title": appTitle,
			"Rows":  enrichedR2Objs,
		})

	}

	// If the value is not cached, get the list of objects from R2 and cache it

	// Get the list of objects from the R2 bucket (without metadata)
	r2Objs, err := h.Scraper.R2.ListObjects()
	if err != nil {
		log.Printf("Error listing objects from R2: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Could not load the list of files. Please try again in a moment.")
	}

	// Enrich the R2 objects with metadata
	R2ObjMetaData, err := h.Scraper.R2.EnrichMetaData(r2Objs)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Could not load the list of files. Please try again in a moment.")
	}

	// Create a slice of Row structs to hold the data for the table
	enrichedR2Objs := make([]enrichedR2Obj, len(R2ObjMetaData))

	// Populate the slice of Row struct with data from the fileNames array
	for i, r2Obj := range R2ObjMetaData {
		uploadDate, err := utils.ParseTime(r2Obj.LastModified)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Could not load the list of files. Please try again in a moment.")
		}
		enrichedR2Objs[i] = enrichedR2Obj{
			FileName:   r2Obj.Key,
			Link:       r2Url + r2Obj.Key,
			UploadedBy: r2Obj.Metadata,
			Date:       uploadDate,
		}
	}

	// Store the encoded byte slice into redis
	h.Scraper.Redis.SetCache("r2StorageObjectsList", enrichedR2Objs)

	return c.Render("downloads", fiber.Map{
		"Title": appTitle,
		"Rows":  enrichedR2Objs,
	})
}
