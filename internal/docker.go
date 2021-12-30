package internal

import (
	"github.com/docker/docker/client"
)

type Docker struct {
	client *client.Client
}

func NewDocker() (*Docker, error) {
	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	return &Docker{
		client: docker,
	}, nil
}

func (d *Docker) Close() error {
	return d.client.Close()
}
