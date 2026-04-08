package containers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	containerImage = "ghcr.io/algo7/tripadvisor-review-scraper/scraper:latest"
	// containerImage = "scraper:latest"
)

var (
	ErrInvalidContainerType = errors.New("invalid container type")
)

type ContainerClient interface {
	ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	CopyFromContainer(ctx context.Context, containerID string, srcPath string) (io.ReadCloser, container.PathStat, error)
	ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
	ContainerStatPath(ctx context.Context, containerID string, path string) (container.PathStat, error)
	ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)
	Close() error
}

type ContainerManager struct {
	Client ContainerClient
	image  string
}

// NewContainerManager creates a new instance of ContainerManager
func NewContainerManager(image string) (*ContainerManager, error) {
	// Create a new Docker API Client
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("fail to create new container manager due to client initialization issues %w", err)
	}
	defer apiClient.Close()

	return &ContainerManager{
		Client: apiClient,
		image:  image,
	}, nil
}

// PullImage pulls the scraper container image
func (c *ContainerManager) PullImage() error {

	reader, err := c.Client.ImagePull(context.Background(), c.image, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("fail to pull the scraper image: %w", err)
	}
	defer reader.Close()

	// Print the progress of the image pull
	_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		return fmt.Errorf("fail to log image pull progress: %w", err)
	}

	return nil
}

// RemoveContainer removes the container with the given ID
func (c *ContainerManager) RemoveContainer(containerID string) error {

	// Remove the container
	err := c.Client.ContainerRemove(context.Background(), containerID, container.RemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   true,
		Force:         true,
	})

	if err != nil {
		return fmt.Errorf("fail to remove finished container: %w", err)
	}

	return nil
}

// TailLog tails the log of the container with the given ID
func (c *ContainerManager) TailLog(containerID string) (io.Reader, error) {

	// Print the logs of the container
	out, err := c.Client.ContainerLogs(context.Background(), containerID, container.LogsOptions{
		ShowStdout: true,
		Details:    true,
		ShowStderr: true,
		Timestamps: false,
		Follow:     true})

	if err != nil {
		return nil, fmt.Errorf("fail to tail container execution log: %w", err)
	}

	// // Docker log uses multiplexed streams to send stdout and stderr in the connection. This function separates them
	// _, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	// utils.ErrorHandler(err)

	return out, nil
}

// CreateContainer creates a container then returns the container ID
func (c *ContainerManager) CreateContainer(containerConfig *container.Config) (string, error) {

	// Create the container. Container.ID contains the ID of the container
	ct, err := c.Client.ContainerCreate(context.Background(),
		containerConfig,
		&container.HostConfig{
			AutoRemove: false, // Cant set to true otherwise the container got deleted before copying the file
		},
		// NetworkConfig
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"scraper_vpn": {
					NetworkID: "scraper_vpn",
				},
			},
		},
		nil, // Platform
		"",  // Container name
	)

	if err != nil {
		return "", fmt.Errorf("fail to create container: %w", err)
	}

	return ct.ID[:12], nil
}

// Container information
type Container struct {
	ContainerID    *string
	TaskOwner      *string
	TargetName     *string
	URL            *string
	IPAddress      *string
	VPNRegion      *string
	ProxySOCKSPort *string
	ProxyHTTPPort  *string
}

// ListContainersByType lists all containers of the given type.
// Available container types:
//   - "scraper": Lists all scraper containers.
//   - "proxy": Lists all proxy containers.
//
// Example:
//
//	scraperContainers := ListContainersByType("scraper")
//	proxyContainers := ListContainersByType("proxy")
func (c *ContainerManager) ListContainersByType(containerType string) ([]Container, error) {

	// List all containers
	containersInfo, err := c.Client.ContainerList(context.Background(), container.ListOptions{All: false})
	if err != nil {
		return nil, fmt.Errorf("fail to list %s containers: %w", containerType, err)
	}

	// Create a slice of Container structs
	containers := []Container{}

	// Iterate through the containers and append them to the slice
	for _, containerInfo := range containersInfo {

		// Extract fields from the container info and map them to the Container struct
		containerID := containerInfo.ID[:12]
		taskOwner := containerInfo.Labels["TaskOwner"]
		targetName := containerInfo.Labels["Target"]
		vpnRegion := containerInfo.Labels["vpn.region"]
		vpnSOCKSPort := containerInfo.Labels["proxy.socks.port"]
		vpnHTTPPort := containerInfo.Labels["proxy.http.port"]

		url := fmt.Sprintf("/logs-viewer?container_id=%s", containerInfo.ID[:12])

		switch containerType {

		// If the container type is "scraper", only list scraper containers
		case "scraper":
			// logic for listing scraper containers
			if taskOwner != "" && taskOwner != "PROXY" {

				containers = append(containers, Container{
					ContainerID: &containerID,
					URL:         &url,
					TaskOwner:   &taskOwner,
					TargetName:  &targetName,
					VPNRegion:   &vpnRegion,
				})
			}

			// If the container type is "proxy", only list proxy containers
		case "proxy":
			if taskOwner == "PROXY" {
				containers = append(containers, Container{
					ContainerID:    &containerID,
					VPNRegion:      &vpnRegion,
					IPAddress:      &containerInfo.NetworkSettings.Networks["scraper_vpn"].IPAddress,
					ProxySOCKSPort: &vpnSOCKSPort,
					ProxyHTTPPort:  &vpnHTTPPort,
				})

			}

		default:
			return nil, ErrInvalidContainerType
		}
	}

	return containers, nil
}

// // getResultCSVSizeInContainer gets the size of the result csv file in the container
// func (c *ContainerManager) getResultCSVSizeInContainer(containerID, filePathInContainer string) error {

// 	// Log the file size in the container
// 	containerFileInfo, err := c.Client.ContainerStatPath(context.Background(), containerID, filePathInContainer)
// 	if err != nil {
// 		return fmt.Errorf("error getting file size in container: %w", err)
// 	} else {
// 		log.Printf("file size in container: %d bytes", containerFileInfo.Size)
// 	}

// 	return nil
// }

// // Scrape creates a container, runs it, tails the log and wait for it to exit, and export the file name
// func (c *ContainerManager) Scrape(uploadIdentifier string, targetName string, containerID string) error {

// 	// Start the container
// 	err := c.Client.ContainerStart(context.Background(), containerID, container.StartOptions{})
// 	if err != nil {
// 		return fmt.Errorf("fail to start container %s: %w", containerID, err)
// 	}

// 	// Wait for the container to exit
// 	statusCh, errCh := c.Client.ContainerWait(context.Background(), containerID, container.WaitConditionNotRunning)

// 	// ContainerWait returns 2 channels. One for the status and one for the wait error (not execution error)
// 	select {
// 	case err := <-errCh:
// 		if err != nil {
// 			return fmt.Errorf("container exited due to %w", err)
// 		}

// 	case status := <-statusCh:
// 		// If the container exited with non-zero status code, remove the container and return an error
// 		if status.StatusCode != 0 {
// 			err := c.RemoveContainer(containerID)
// 			if err != nil {
// 				return fmt.Errorf("fail to container %s: %w", containerID, err)
// 			}
// 			return nil
// 		}
// 	}

// 	// The file path in the container
// 	filePathInContainer := "reviews.csv"

// 	// Get the file size in the container
// 	err = c.getResultCSVSizeInContainer(containerID, filePathInContainer)
// 	if err != nil {
// 		return fmt.Errorf("fail to get csv size in container %s: %w", containerID, err)
// 	}

// 	// Read the file from the container as a reader interface of a tar stream
// 	fileReader, _, err := c.Client.CopyFromContainer(context.Background(), containerID, filePathInContainer)
// 	if err != nil {
// 		return fmt.Errorf("fail to copy file from container %s: %w", containerID, err)
// 	}

// 	// Generate a random file prefix
// 	fileSuffix := utils.GenerateUUID()

// 	// Write the file to the host
// 	exportedFileName, err := utils.WriteToFileFromTarStream(targetName, fileSuffix, fileReader)
// 	if err != nil {
// 		return fmt.Errorf("fail to write file to host: %w", err)
// 	}

// 	// Read the exported csv file
// 	file, err := utils.ReadFromFile(exportedFileName)
// 	if err != nil {
// 		return fmt.Errorf("fail to read exported file %s: %w", exportedFileName, err)
// 	}

// 	// Upload the file to R2
// 	err = utils.R2UploadObject(exportedFileName, uploadIdentifier, file)
// 	if err != nil {
// 		return fmt.Errorf("fail to upload file %s to R2: %w", exportedFileName, err)
// 	}

// 	// Remove the container
// 	err = c.RemoveContainer(containerID)
// 	if err != nil {
// 		return fmt.Errorf("fail to remove container %s after finishing uploading file: %w", containerID, err)
// 	}

// 	return nil
// }
