package scrape

import (
	"context"
	"fmt"
	"log"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/containers"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/database"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/storage"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/utils"
	"github.com/docker/docker/api/types/container"
)

// ProxyContainer information
type ProxyContainer struct {
	ContainerID  string
	ProxyAddress string
	VPNRegion    string
}

type Scraper struct {
	cm    *containers.ContainerManager
	r2    *storage.R2Service
	redis *database.RedisClient
}

// NewScraper creates a new Scraper instance with the given ContainerManager, R2Service, and RedisClient
func NewScraper(cm *containers.ContainerManager, r2 *storage.R2Service, redis *database.RedisClient) *Scraper {
	return &Scraper{
		cm:    cm,
		r2:    r2,
		redis: redis,
	}
}

// Scrape creates a container, runs it, tails the log and wait for it to exit, and export the file name
func (s *Scraper) Scrape(uploadIdentifier string, targetName string, containerID string) error {

	// Start the container
	err := s.cm.Client.ContainerStart(context.Background(), containerID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("fail to start container %s: %w", containerID, err)
	}

	// Wait for the container to exit
	statusCh, errCh := s.cm.Client.ContainerWait(context.Background(), containerID, container.WaitConditionNotRunning)

	// ContainerWait returns 2 channels. One for the status and one for the wait error (not execution error)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("container exited due to %w", err)
		}

	case status := <-statusCh:
		// If the container exited with non-zero status code, remove the container and return an error
		if status.StatusCode != 0 {
			err := s.cm.RemoveContainer(containerID)
			if err != nil {
				return fmt.Errorf("fail to container %s: %w", containerID, err)
			}
			return nil
		}
	}

	// The file path in the container
	filePathInContainer := "reviews.csv"

	// Get the file size in the container
	// 	// Log the file size in the container
	containerFileInfo, err := s.cm.Client.ContainerStatPath(context.Background(), containerID, filePathInContainer)
	if err != nil {
		return fmt.Errorf("error getting csv file size in container: %w", err)
	} else {
		log.Printf("file size in container: %d bytes", containerFileInfo.Size)
	}

	// Read the file from the container as a reader interface of a tar stream
	fileReader, _, err := s.cm.Client.CopyFromContainer(context.Background(), containerID, filePathInContainer)
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
	err = s.r2.UploadObject(exportedFileName, uploadIdentifier, file)
	if err != nil {
		return fmt.Errorf("fail to upload file %s to R2: %w", exportedFileName, err)
	}

	// Remove the container
	err = s.cm.RemoveContainer(containerID)
	if err != nil {
		return fmt.Errorf("fail to remove container %s after finishing uploading file: %w", containerID, err)
	}

	return nil
}

// AcquireProxyContainer acquires a lock on a proxy container and returns its ID
func AcquireProxyContainer(s *Scraper) (ProxyContainer, error) {
	availableProxies, err := s.cm.ListContainersByType("proxy")
	if err != nil {
		return ProxyContainer{}, fmt.Errorf("failed to list proxy containers: %w", err)
	}

	for _, proxy := range availableProxies {
		lockKey := "proxy-usage:" + *proxy.ContainerID
		lockSuccess := s.redis.SetLock(lockKey)

		if lockSuccess && proxy.ProxySOCKSPort != nil && proxy.IPAddress != nil {
			return ProxyContainer{
				ContainerID:  *proxy.ContainerID,
				VPNRegion:    *proxy.VPNRegion,
				ProxyAddress: fmt.Sprintf("socks5://%s:%s", *proxy.IPAddress, *proxy.ProxySOCKSPort),
			}, nil
		}
		// If the lock is not successful, try the next proxy container
	}

	// If no proxy container could be locked, return an empty string
	return ProxyContainer{}, fmt.Errorf("no available proxy container")
}

// ReleaseProxyContainer releases the lock on a proxy container
func ReleaseProxyContainer(s *Scraper, containerID string) {
	lockKey := "proxy-usage:" + containerID
	log.Printf("Releasing lock on proxy container %s", lockKey)
	s.redis.ReleaseLock(lockKey)
}
