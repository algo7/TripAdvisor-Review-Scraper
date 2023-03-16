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
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	// Pull the image
	reader, err := cli.ImagePull(ctx, "ghcr.io/algo7/tripadvisor-review-scraper/scrap:latest", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	// Print the progress of the image pull
	io.Copy(os.Stdout, reader)

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

	if err != nil {
		panic(err)
	}

	// Start the container
	cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})

	if err != nil {
		panic(err)
	}

	// Print the logs of the container
	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, Follow: true})
	if err != nil {
		panic(err)
	}
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	tarReader, _, err := cli.CopyFromContainer(ctx, resp.ID, "/puppeteer/reviews/All.csv")

	if err != nil {
		panic(err)
	}

	utils.WriteToFile("All.csv", tarReader)

	// Wait for the container to exit
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:

	}
}
