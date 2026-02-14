package containers

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
)

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
