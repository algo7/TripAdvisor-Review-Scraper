package containers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/database"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/utils"
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
	Close() error
}

type ContainerManager struct {
	client ContainerClient
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
		client: apiClient,
		image:  image,
	}, nil
}

// PullImage pulls the scraper container image
func (c *ContainerManager) PullImage() error {

	reader, err := c.client.ImagePull(context.Background(), c.image, image.PullOptions{})
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
	err := c.client.ContainerRemove(context.Background(), containerID, container.RemoveOptions{
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
	out, err := c.client.ContainerLogs(context.Background(), containerID, container.LogsOptions{
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
func (c *ContainerManager) CreateContainer(containerConfig *container.Config, networkConfig *network.NetworkingConfig) (string, error) {

	// Create the container. Container.ID contains the ID of the container
	ct, err := c.client.ContainerCreate(context.Background(),
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

// Scrape creates a container, runs it, tails the log and wait for it to exit, and export the file name
func (c *ContainerManager) Scrape(uploadIdentifier string, targetName string, containerID string) error {

	// Start the container
	err := c.client.ContainerStart(context.Background(), containerID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("fail to start container %s: %w", containerID, err)
	}

	// Wait for the container to exit
	statusCh, errCh := c.client.ContainerWait(context.Background(), containerID, container.WaitConditionNotRunning)

	// ContainerWait returns 2 channels. One for the status and one for the wait error (not execution error)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("container exited due to %w", err)
		}

	case status := <-statusCh:
		// If the container exited with non-zero status code, remove the container and return an error
		if status.StatusCode != 0 {
			err := c.RemoveContainer(containerID)
			if err != nil {
				return fmt.Errorf("fail to container %s: %w", containerID, err)
			}
			return nil
		}
	}

	// The file path in the container
	filePathInContainer := "reviews.csv"

	// Get the file size in the container
	err = c.getResultCSVSizeInContainer(containerID, filePathInContainer)
	if err != nil {
		return fmt.Errorf("fail to get csv size in container %s: %w", containerID, err)
	}

	// Read the file from the container as a reader interface of a tar stream
	fileReader, _, err := c.client.CopyFromContainer(context.Background(), containerID, filePathInContainer)
	if err != nil {
		return fmt.Errorf("fail to copy file from container %s: %w", containerID, err)
	}

	// Generate a random file prefix
	fileSuffix := utils.GenerateUUID()

	// Write the file to the host
	exportedFileName, err := utils.WriteToFileFromTarStream(targetName, fileSuffix, fileReader)
	if err != nil {
		return fmt.Errorf("fail to write file to host: %w", err)
	}

	// Read the exported csv file
	file, err := utils.ReadFromFile(exportedFileName)
	if err != nil {
		return fmt.Errorf("fail to read exported file %s: %w", exportedFileName, err)
	}

	// Upload the file to R2
	err = utils.R2UploadObject(exportedFileName, uploadIdentifier, file)
	if err != nil {
		return fmt.Errorf("fail to upload file %s to R2: %w", exportedFileName, err)
	}

	// Remove the container
	err = c.RemoveContainer(containerID)
	if err != nil {
		return fmt.Errorf("fail to remove container %s after finishing uploading file: %w", containerID, err)
	}

	return nil
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
	containersInfo, err := c.client.ContainerList(context.Background(), container.ListOptions{All: false})
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

// ProxyContainer information
type ProxyContainer struct {
	ContainerID  string
	ProxyAddress string
	VPNRegion    string
}

// AcquireProxyContainer acquires a lock on a proxy container and returns its ID
func AcquireProxyContainer() ProxyContainer {
	availableProxies := ListContainersByType("proxy")

	for _, proxy := range availableProxies {
		lockKey := "proxy-usage:" + *proxy.ContainerID
		lockSuccess := database.SetLock(lockKey)

		if lockSuccess && proxy.ProxySOCKSPort != nil && proxy.IPAddress != nil {
			return ProxyContainer{
				ContainerID:  *proxy.ContainerID,
				VPNRegion:    *proxy.VPNRegion,
				ProxyAddress: fmt.Sprintf("socks5://%s:%s", *proxy.IPAddress, *proxy.ProxySOCKSPort),
			}
		}
		// If the lock is not successful, try the next proxy container
	}

	// If no proxy container could be locked, return an empty string
	return ProxyContainer{}
}

// ReleaseProxyContainer releases the lock on a proxy container
func ReleaseProxyContainer(containerID string) {
	lockKey := "proxy-usage:" + containerID
	log.Printf("Releasing lock on proxy container %s", lockKey)
	database.ReleaseLock(lockKey)
}

// getResultCSVSizeInContainer gets the size of the result csv file in the container
func (c *ContainerManager) getResultCSVSizeInContainer(containerID, filePathInContainer string) error {

	// Log the file size in the container
	containerFileInfo, err := c.client.ContainerStatPath(context.Background(), containerID, filePathInContainer)
	if err != nil {
		return fmt.Errorf("error getting file size in container: %w", err)
	} else {
		log.Printf("file size in container: %d bytes", containerFileInfo.Size)
	}

	return nil
}
