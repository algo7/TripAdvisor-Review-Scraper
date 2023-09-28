package containers

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/utils"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// initializeDockerClient initialize a new docker api client
func initializeDockerClient() *client.Client {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	return cli
}

// pullImage pulls the given image from a registry
func pullImage(image string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	defer cli.Close()

	reader, err := cli.ImagePull(context.Background(), image, types.ImagePullOptions{})
	utils.ErrorHandler(err)
	defer reader.Close()

	// Print the progress of the image pull
	_, err = io.Copy(os.Stdout, reader)
	utils.ErrorHandler(err)
}

// removeContainer removes the container with the given ID
func removeContainer(containerId string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	defer cli.Close()

	// Remove the container
	err = cli.ContainerRemove(context.Background(), containerId, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
	utils.ErrorHandler(err)
}

// CreateContainer creates a container then returns the container ID
func CreateContainer(hotelName string, hotelUrl string, uploadIdentifier string) string {

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	defer cli.Close()

	// Create the container. Container.ID contains the ID of the container
	Container, err := cli.ContainerCreate(context.Background(),
		&container.Config{
			Image: "ghcr.io/algo7/tripadvisor-review-scraper/scraper:latest",
			Labels: map[string]string{
				"TaskOwner": uploadIdentifier,
				"Hotel":     hotelName,
			},
			// Env vars required by the js scraper containers
			Env: []string{
				"CONCURRENCY=1",
				"SCRAPE_MODE=HOTEL",
				"HOTEL_NAME=" + hotelName,
				"IS_PROVISIONER=true",
				"HOTEL_URL=" + hotelUrl,
			},
			Tty: true,
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

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	defer cli.Close()

	// Determine if the current process is running inside a container
	isContainer := os.Getenv("IS_CONTAINER")

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		All: false, // Only running containers
	})
	utils.ErrorHandler(err)

	if isContainer == "" {
		return len(containers)
	}
	// 21 otherwise the current process and redis will also be counted as a running container
	return len(containers) - 2
}

// tailLog tails the log of the container with the given ID
func TailLog(containerId string) io.Reader {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	defer cli.Close()

	// Print the logs of the container
	out, err := cli.ContainerLogs(context.Background(), containerId, types.ContainerLogsOptions{
		ShowStdout: true,
		Details:    true,
		ShowStderr: false,
		Timestamps: false,
		Follow:     true})
	utils.ErrorHandler(err)

	// // Docker log uses multiplexed streams to send stdout and stderr in the connection. This function separates them
	// _, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	// utils.ErrorHandler(err)

	return out
}

type Container struct {
	ID        string
	TaskOwner string
	HotelName string
}

// ListContainers lists all the containers and return the container IDs
func ListContainers() []Container {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	defer cli.Close()

	containersInfo, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	utils.ErrorHandler(err)

	// Map container list result into custom container struct
	containers := make([]Container, len(containersInfo))

	for i, containerInfo := range containersInfo {
		containers[i] = Container{
			ID:        containerInfo.ID,
			TaskOwner: containerInfo.Labels["TaskOwner"],
			HotelName: containerInfo.Labels["Hotel"],
		}
	}

	return containers
}

// GetResultCSVSizeInContainer gets the size of the result csv file in the container
func getResultCSVSizeInContainer(containerId, filePathInContainer string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	defer cli.Close()

	// Log the file size in the container
	containerFileInfo, err := cli.ContainerStatPath(context.Background(), containerId, filePathInContainer)
	if err == nil {
		log.Printf("File size in container: %d bytes", containerFileInfo.Size)
	} else {
		log.Printf("Error getting file size in container: %v", err)
	}
}
