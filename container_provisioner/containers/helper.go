package containers

import (
	"container_provisioner/utils"
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

var (
	ctx = context.Background()
	cli = initializeDockerClient()
)

func initializeDockerClient() *client.Client {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	return cli
}

// CountRunningContainer lists the number of running containers
func CountRunningContainer() int {

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		All: false, // Only running containers
	})
	utils.ErrorHandler(err)

	return len(containers)
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

// tailLog tails the log of the container with the given ID
func tailLog(containerId string) {

	// Print the logs of the container
	out, err := cli.ContainerLogs(ctx, containerId, types.ContainerLogsOptions{ShowStdout: true, Follow: true})
	utils.ErrorHandler(err)

	// Docker log uses multiplexed streams to send stdout and stderr in the connection. This function separates them
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
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
