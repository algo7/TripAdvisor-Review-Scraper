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

	// Create the container. resp.ID contains the ID of the container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "test:latest",
		Env: []string{
			"CONCURRENCY=1",
			"SCRAPE_MODE=HOTEL",
			"HOTEL_NAME=BRO",
			"HOTEL_URL=https://www.tripadvisor.com/Hotel_Review-g188107-d199124-Reviews-Hotel_Des_Voyageurs-Lausanne_Canton_of_Vaud.html"},
		Tty: false,
	}, nil, nil, nil, "")
	utils.ErrorHandler(err)

	// Start the container
	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	utils.ErrorHandler(err)

	// Print the logs of the container
	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, Follow: true})
	utils.ErrorHandler(err)

	// Docker log uses multiplexed streams to send stdout and stderr in the connection. This function separates them
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	utils.ErrorHandler(err)

	fileReader, _, err := cli.CopyFromContainer(ctx, resp.ID, "/puppeteer/reviews/All.csv")
	utils.ErrorHandler(err)

	err = utils.WriteToFile("Reviews.csv", fileReader)
	utils.ErrorHandler(err)

	// Wait for the container to exit
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	// ContainerWait returns 2 channels. One for the status and one for the error
	select {
	case err := <-errCh:
		utils.ErrorHandler(err)
	case <-statusCh:
	}
}
