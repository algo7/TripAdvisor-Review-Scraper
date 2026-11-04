package scrape

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/containers"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/database"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/storage"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/utils"
	"github.com/google/uuid"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

// ProxyContainer information
type ProxyContainer struct {
	ContainerID  string
	ProxyAddress string
	VPNRegion    string
	// lockToken identifies the scrape that holds the lock on this proxy container
	lockToken string
}

type Scraper struct {
	CM    *containers.ContainerManager
	R2    *storage.R2Service
	Redis *database.RedisClient
}

// NewScraper creates a new Scraper instance with the given ContainerManager, R2Service, and RedisClient
func NewScraper(cm *containers.ContainerManager, r2 *storage.R2Service, redis *database.RedisClient) *Scraper {
	return &Scraper{
		CM:    cm,
		R2:    r2,
		Redis: redis,
	}
}

// Scrape creates a container, runs it, tails the log and wait for it to exit, and export the file name
func (s *Scraper) Scrape(uploadIdentifier string, targetName string, containerID string) error {

	// Start the container
	_, err := s.CM.Client.ContainerStart(context.Background(), containerID, client.ContainerStartOptions{})
	if err != nil {
		return fmt.Errorf("fail to start container %s: %w", containerID, err)
	}

	// Wait for the container to exit
	wait := s.CM.Client.ContainerWait(context.Background(), containerID, client.ContainerWaitOptions{Condition: container.WaitConditionNotRunning})

	// ContainerWait returns 2 channels. One for the status and one for the wait error (not execution error)
	select {
	case err := <-wait.Error:
		if err != nil {
			return fmt.Errorf("container exited due to %w", err)
		}

	case status := <-wait.Result:
		// The scrape failed. Keep the container, so its logs and file system can be inspected.
		if status.StatusCode != 0 {
			return fmt.Errorf("scraper container %s exited with status code %d", containerID, status.StatusCode)
		}
	}

	// The file path in the container
	filePathInContainer := "reviews.csv"

	// Get the file size in the container
	// 	// Log the file size in the container
	containerFileInfo, err := s.CM.Client.ContainerStatPath(context.Background(), containerID, client.ContainerStatPathOptions{Path: filePathInContainer})
	if err != nil {
		return fmt.Errorf("error getting csv file size in container: %w", err)
	} else {
		log.Printf("file size in container: %d bytes", containerFileInfo.Stat.Size)
	}

	// Read the file from the container as a reader interface of a tar stream
	copied, err := s.CM.Client.CopyFromContainer(context.Background(), containerID, client.CopyFromContainerOptions{SourcePath: filePathInContainer})
	if err != nil {
		return fmt.Errorf("fail to copy file from container %s: %w", containerID, err)
	}
	fileReader := copied.Content

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
	err = s.R2.UploadObject(exportedFileName, uploadIdentifier, file)
	if err != nil {
		return fmt.Errorf("fail to upload file %s to R2: %w", exportedFileName, err)
	}

	// Remove the container
	err = s.CM.RemoveContainer(containerID)
	if err != nil {
		return fmt.Errorf("fail to remove container %s after finishing uploading file: %w", containerID, err)
	}

	return nil
}

// proxyLockTTL is how long a proxy lock lasts without being renewed. It only runs out when
// the provisioner stops renewing it, e.g. after a crash, so the proxy becomes free again.
var proxyLockTTL = 2 * time.Minute

// proxyLockRenewInterval is how often a running scrape renews its proxy lock. It is much shorter
// than proxyLockTTL, so a few renewals can fail before the lock expires.
var proxyLockRenewInterval = 30 * time.Second

func proxyLockKey(containerID string) string {
	return "proxy-usage:" + containerID
}

// AcquireProxyContainer locks a free proxy container for one scrape and returns it
func (s *Scraper) AcquireProxyContainer() (ProxyContainer, error) {
	availableProxies, err := s.CM.ListContainersByType("proxy")
	if err != nil {
		return ProxyContainer{}, fmt.Errorf("failed to list proxy containers: %w", err)
	}

	for _, proxy := range availableProxies {
		// Skip proxies that can't be used before locking them, so no lock is left behind
		if *proxy.ProxySOCKSPort == "" || *proxy.IPAddress == "" {
			continue
		}

		// A token unique to this scrape, so only this scrape can renew or release the lock
		lockToken := uuid.NewString()
		acquired, err := s.Redis.AcquireLock(proxyLockKey(*proxy.ContainerID), lockToken, proxyLockTTL)
		if err != nil {
			return ProxyContainer{}, err
		}

		if acquired {
			return ProxyContainer{
				ContainerID:  *proxy.ContainerID,
				VPNRegion:    *proxy.VPNRegion,
				ProxyAddress: fmt.Sprintf("socks5://%s:%s", *proxy.IPAddress, *proxy.ProxySOCKSPort),
				lockToken:    lockToken,
			}, nil
		}
		// The proxy is used by another scrape, try the next one
	}

	return ProxyContainer{}, fmt.Errorf("no available proxy container")
}

// HoldProxyContainer renews the lock on the proxy container until ctx is cancelled,
// so the lock lasts as long as the scrape does
func (s *Scraper) HoldProxyContainer(ctx context.Context, proxy ProxyContainer) {
	ticker := time.NewTicker(proxyLockRenewInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			renewed, err := s.Redis.RenewLock(proxyLockKey(proxy.ContainerID), proxy.lockToken, proxyLockTTL)
			if err != nil {
				// Keep trying, the lock only expires after several failed renewals
				log.Printf("fail to renew lock on proxy container %s: %v", proxy.ContainerID, err)
				continue
			}
			if !renewed {
				log.Printf("lost the lock on proxy container %s, another scrape may be using it", proxy.ContainerID)
				return
			}
		}
	}
}

// ReleaseProxyContainer releases the lock on a proxy container, if this scrape still holds it
func (s *Scraper) ReleaseProxyContainer(proxy ProxyContainer) {
	released, err := s.Redis.ReleaseLock(proxyLockKey(proxy.ContainerID), proxy.lockToken)
	if err != nil {
		log.Printf("fail to release lock on proxy container %s: %v", proxy.ContainerID, err)
		return
	}
	if !released {
		log.Printf("lock on proxy container %s was no longer held by this scrape", proxy.ContainerID)
		return
	}
	log.Printf("released lock on proxy container %s", proxy.ContainerID)
}
