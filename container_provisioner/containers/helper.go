package containers

import (
	"container_provisioner/utils"
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

var (
	ctx = context.Background()
	cli = initializeDockerClient()
)

// initializeDockerClient initialize a new docker api client
func initializeDockerClient() *client.Client {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	return cli
}

// pullImage pulls the given image from a registry
func pullImage(image string) {

	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	utils.ErrorHandler(err)
	defer reader.Close()

	// Print the progress of the image pull
	_, err = io.Copy(os.Stdout, reader)
	utils.ErrorHandler(err)
}

// removeContainer removes the container with the given ID
func removeContainer(containerId string) {

	// Remove the container
	err := cli.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
	utils.ErrorHandler(err)
}

// CreateContainer creates a container then returns the container ID
func CreateContainer(hotelName string, hotelUrl string) string {

	// Create the container. Container.ID contains the ID of the container
	Container, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image: "scrape:latest",
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

	return Container.ID
}

// CountRunningContainer lists the number of running containers
func CountRunningContainer() int {

	// Determine if the current process is running inside a container
	isContainer := os.Getenv("IS_CONTAINER")

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		All: false, // Only running containers
	})
	utils.ErrorHandler(err)

	if isContainer == "" {
		return len(containers)
	}
	// -1 otherwise the current process will also be counted as a running container
	return len(containers) - 1
}

// tailLog tails the log of the container with the given ID
func TailLog(containerId string) io.Reader {

	// Print the logs of the container
	out, err := cli.ContainerLogs(ctx, containerId, types.ContainerLogsOptions{ShowStdout: true, Follow: true})
	utils.ErrorHandler(err)

	// // Docker log uses multiplexed streams to send stdout and stderr in the connection. This function separates them
	// _, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	// utils.ErrorHandler(err)

	return out
}

// ListContainers lists all the containers and return the container IDs
func ListContainers() []string {

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	utils.ErrorHandler(err)

	// Slice to hold container ids
	containerIds := []string{}

	for _, container := range containers {
		containerIds = append(containerIds, container.ID)
	}

	return containerIds
}
