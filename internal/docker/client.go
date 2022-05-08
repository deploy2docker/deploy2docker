package docker

import (
	"context"

	"github.com/deploy2docker/deploy2docker/internal/config"
	"github.com/deploy2docker/deploy2docker/internal/remote"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type Docker struct {
	client *client.Client
}

// New docker client from ssh connection
func NewDockerClient() (*Docker, error) {
	options := client.WithHost(remote.LocalDockerHost)
	docker, err := client.NewClientWithOpts(options)
	if err != nil {
		return nil, err
	}

	return &Docker{
		client: docker,
	}, nil
}

// Close the docker client connection
func (d *Docker) Close() error {
	return d.client.Close()
}

// Ping returns true if the Docker client is actively running
func (d *Docker) Ping(ctx context.Context) bool {
	ping, err := d.client.Ping(ctx)
	if err != nil {
		logrus.Errorln(err)
		return false
	}

	if ping.APIVersion == "" {
		return false
	}

	return true
}

func (d *Docker) Deploy(ctx context.Context, config *config.Config) error {

	// deploy the docker image

	for _, service := range config.Services {

		// create the docker network
		for _, network := range service.Networks {
			if _, err := d.GetNetworkByName(network); err != nil {
				if _, err := d.CreateBridgeNetwork(ctx, network); err != nil {
					return err
				}
			}
		}

		// stop the container if it exists
		if container, err := d.GetContainerByName(ctx, service.Name); err == nil {
			if err := d.StopContainer(ctx, container.ID); err != nil {
				return err
			}

			if err := d.RemoveContainer(ctx, container.ID); err != nil {
				return err
			}
		}

		// create the container
		container, err := d.CreateContainer(ctx, service)
		if err != nil {
			return err
		}

		// start the container
		if err := d.StartContainer(ctx, container.ID); err != nil {
			return err
		}

		// wait for the container to start
		if err := d.ContainerLogs(ctx, container.ID); err != nil {
			return err
		}

		logrus.Infof("Deployed %s", service.Name)
	}

	return nil
}
