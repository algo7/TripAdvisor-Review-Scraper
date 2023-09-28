package containers

import (
	"context"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/utils"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// Scrape creates a container, runs it, tails the log and wait for it to exit, and export the file name
func Scrape(uploadIdentifier string, hotelName string, containerId string) {

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	defer cli.Close()

	// Start the container
	err = cli.ContainerStart(context.Background(), containerId, types.ContainerStartOptions{})
	utils.ErrorHandler(err)

	// Wait for the container to exit
	statusCh, errCh := cli.ContainerWait(context.Background(), containerId, container.WaitConditionNotRunning)

	// ContainerWait returns 2 channels. One for the status and one for the wait error (not execution error)
	select {
	case err := <-errCh:
		utils.ErrorHandler(err)

	case status := <-statusCh:
		// If the container exited with non-zero status code, remove the container and return an error
		if status.StatusCode != 0 {
			removeContainer(containerId)
			return
		}
	}

	// The file path in the container
	filePathInContainer := "/puppeteer/reviews/All.csv"

	// Get the file size in the container
	getResultCSVSizeInContainer(containerId, filePathInContainer)

	// Read the file from the container as a reader interface of a tar stream
	fileReader, _, err := cli.CopyFromContainer(context.Background(), containerId, filePathInContainer)
	utils.ErrorHandler(err)

	// Generate a random file prefix
	fileSuffix := utils.GenerateUUID()

	// Write the file to the host
	exportedFileName := utils.WriteToFileFromTarStream(hotelName, fileSuffix, fileReader)

	// Read the exported csv file
	file := utils.ReadFromFile(exportedFileName)

	// Upload the file to R2
	utils.R2UploadObject(exportedFileName, uploadIdentifier, file)

	// Remove the container
	removeContainer(containerId)
}
