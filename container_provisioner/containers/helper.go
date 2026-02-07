package containers

import (
	"context"
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
)

const (
	containerImage = "ghcr.io/algo7/tripadvisor-review-scraper/scraper:latest"
	// containerImage = "scraper:latest"
)

type ContainerManager struct {
	client *client.Client
	image  string
	config *container.Config
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

// ContainerConfigGenerator generates the container config depending on the scrape target
func (c *ContainerManager) ContainerConfigGenerator(
	locationURL string, locationName string, uploadIdentifier string,
	proxyAddress string, vpnRegion string) *container.Config {

	return &container.Config{
		Image: containerImage,
		Labels: map[string]string{
			"TaskOwner":  uploadIdentifier,
			"Target":     locationName,
			"vpn.region": vpnRegion,
			"TargetName": locationName,
		},
		// Env vars required by the scraper containers
		Env: []string{
			fmt.Sprintf("LOCATION_URL=%s", locationURL),
			fmt.Sprintf("PROXY_HOST=%s", proxyAddress),
		},
		Tty: true,
	}
}

// CreateContainer creates a container then returns the container ID
func (c *ContainerManager) CreateContainer(containerConfig *container.Config) (string, error) {

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
func (c *ContainerManager) ListContainersByType(containerType string) []Container {

	// List all containers
	containersInfo, err := c.client.ContainerList(context.Background(), container.ListOptions{All: false})
	utils.ErrorHandler(err)

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
			utils.ErrorHandler(fmt.Errorf("Invalid container type"))
		}
	}

	return containers
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
func getResultCSVSizeInContainer(containerID, filePathInContainer string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	defer cli.Close()

	// Log the file size in the container
	containerFileInfo, err := cli.ContainerStatPath(context.Background(), containerID, filePathInContainer)
	if err == nil {
		log.Printf("File size in container: %d bytes", containerFileInfo.Size)
	} else {
		log.Printf("Error getting file size in container: %v", err)
	}
}
