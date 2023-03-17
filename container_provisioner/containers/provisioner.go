package containers

import (
	"container_provisioner/utils"
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// Provision creates a container, runs it, tails the log and wait for it to exit
func Provision() {
	ctx := context.Background()

	// Connect to the Docker daemon
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	defer cli.Close()

	// Pull the image
	reader, err := cli.ImagePull(ctx, "ghcr.io/algo7/tripadvisor-review-scraper/scrap:latest", types.ImagePullOptions{})
	utils.ErrorHandler(err)
	defer reader.Close()

	// Print the progress of the image pull
	_, err = io.Copy(os.Stdout, reader)
	utils.ErrorHandler(err)

	// Create the container. Container.ID contains the ID of the container
	Container, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image: "test:latest",
			Env: []string{
				"CONCURRENCY=1",
				"SCRAPE_MODE=HOTEL",
				"HOTEL_NAME=BRO",
				"HOTEL_URL=https://www.tripadvisor.com/Hotel_Review-g188107-d199124-Reviews-Hotel_Des_Voyageurs-Lausanne_Canton_of_Vaud.html"},
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
	err = cli.ContainerStart(ctx, Container.ID, types.ContainerStartOptions{})
	utils.ErrorHandler(err)

	// Print the logs of the container
	out, err := cli.ContainerLogs(ctx, Container.ID, types.ContainerLogsOptions{ShowStdout: true, Follow: true})
	utils.ErrorHandler(err)

	// Docker log uses multiplexed streams to send stdout and stderr in the connection. This function separates them
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	utils.ErrorHandler(err)

	// Read the file from the container as a reader interface of a tar stream
	fileReader, _, err := cli.CopyFromContainer(ctx, Container.ID, "/puppeteer/reviews/All.csv")
	utils.ErrorHandler(err)

	// Write the file to the host
	err = utils.WriteToFile("Reviews.csv", fileReader)
	utils.ErrorHandler(err)

	// Wait for the container to exit
	statusCh, errCh := cli.ContainerWait(ctx, Container.ID, container.WaitConditionNotRunning)

	// ContainerWait returns 2 channels. One for the status and one for the error
	select {
	case err := <-errCh:
		utils.ErrorHandler(err)
	case <-statusCh:
	}

	// Remove the container
	err = cli.ContainerRemove(ctx, Container.ID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
	utils.ErrorHandler(err)
}