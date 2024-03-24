package containers

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// Container struct repesents a single container
type Container struct {
	cli           *client.Client
	config        *container.Config
	hostConfig    *container.HostConfig
	networkConfig *network.NetworkingConfig
	removeOptions *container.RemoveOptions
	ID            string
}

// New function returns a new container struct
func New() (*Container, error) {

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &Container{
		cli:    cli,
		config: &container.Config{},
		hostConfig: &container.HostConfig{
			// Otherwise the container will be removed before we can copy the file
			AutoRemove: false,
		},
		networkConfig: &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"scraper_vpn": {
					NetworkID: "scraper_vpn",
				},
			},
		},
		removeOptions: &container.RemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		},
	}, nil
}

// PullImage pulls the image with the given name (name:tag)
func (c *Container) PullImage(imageName string) error {

	reader, err := c.cli.ImagePull(context.TODO(), imageName, image.PullOptions{})
	defer reader.Close()

	if err != nil {
		return err
	}

	// Print the progress
	_, err = io.Copy(os.Stdout, reader)

	if err != nil {
		return err
	}

	log.Printf("Image %s pulled", imageName)

	return nil
}

// CreateContainer creates a container base on the container struct
func (c *Container) CreateContainer() error {

	resp, err := c.cli.ContainerCreate(context.TODO(), c.config, c.hostConfig, c.networkConfig, nil, "")
	if err != nil {
		return err
	}

	// Set the container ID
	c.ID = resp.ID[:12]

	log.Printf("Container %s created successfully.", c.ID)

	return nil
}

// Remove container removes the container base on the container struct
func (c *Container) RemoveContainer() error {

	err := c.cli.ContainerRemove(context.TODO(), c.ID, *c.removeOptions)
	if err != nil {
		return err
	}

	log.Printf("Container %s[%s] removed sucessfully.", c.config.Hostname, c.ID)

	return nil
}
