package containers

import (
	"container_provisioner/utils"
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// Provision creates a container, runs it, tails the log and wait for it to exit, and export the file name
func Provision(fileSuffix string, uploadIdentifier string, hotelUrl string) {

	// Dedicated context and client for each function call
	ctx1 := context.Background()
	cli1, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	defer cli1.Close()

	// Get the hotel name from the URL
	hotelName := utils.GetHotelNameFromURL(hotelUrl)

	// Create the container. Container.ID contains the ID of the container
	Container, err := cli1.ContainerCreate(ctx1,
		&container.Config{
			Image: "ghcr.io/algo7/tripadvisor-review-scraper/scrape:latest",
			// Env vars required by the js scraper containers
			Env: []string{
				"CONCURRENCY=1",
				"SCRAPE_MODE=HOTEL",
				"HOTEL_NAME=" + hotelName,
				"IS_PROVISIONER=true",
				"HOTEL_URL=" + hotelUrl,
			},
		},
		&container.HostConfig{
			AutoRemove: false, // Cant set to true otherwise the container got deleted before copying the file
		},
		nil, // NetworkConfig
		nil, // Platform
		"",  // Container name
	)
	utils.ErrorHandler(err)

	// Start the container
	err = cli1.ContainerStart(ctx1, Container.ID, types.ContainerStartOptions{})
	utils.ErrorHandler(err)

	// Wait for the container to exit
	statusCh, errCh := cli1.ContainerWait(ctx1, Container.ID, container.WaitConditionNotRunning)

	// ContainerWait returns 2 channels. One for the status and one for the wait error (not execution error)
	select {
	case err := <-errCh:
		utils.ErrorHandler(err)

	case status := <-statusCh:
		// If the container exited with non-zero status code, remove the container and return an error
		if status.StatusCode != 0 {
			removeContainer(Container.ID)
		}
	}

	// The file path in the container
	filePathInContainer := fmt.Sprintf("/puppeteer/reviews/0_%s.csv", hotelName)

	// Read the file from the container as a reader interface of a tar stream
	fileReader, _, err := cli1.CopyFromContainer(ctx1, Container.ID, filePathInContainer)
	utils.ErrorHandler(err)

	// Write the file to the host
	exportedFileName := utils.WriteToFileFromTarStream(fileSuffix, fileReader)

	// Read the exported csv file
	file := utils.ReadFromFile(exportedFileName)

	// Upload the file to R2
	utils.R2UploadObject(exportedFileName, uploadIdentifier, file)

	// Remove the container
	removeContainer(Container.ID)
}
