package docker

import (
	"context"

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
