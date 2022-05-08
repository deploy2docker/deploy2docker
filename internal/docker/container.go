package docker

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/deploy2docker/deploy2docker/internal/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
)

func (c *Docker) GetContainerByName(ctx context.Context, container string) (*types.Container, error) {
	containers, err := c.client.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return nil, err
	}

	for _, c := range containers {
		for _, name := range c.Names {
			n := "/" + container
			if name == n {
				return &c, nil
			}
		}
	}

	return nil, fmt.Errorf("container %s not found", container)
}

// CreateContainer creates a docker container from a docker config
func (c *Docker) CreateContainer(ctx context.Context, service config.Service) (*container.ContainerCreateCreatedBody, error) {

	// create the docker network
	for _, network := range service.Networks {
		if _, err := c.GetNetworkByName(network); err != nil {
			if _, err := c.CreateBridgeNetwork(ctx, network); err != nil {
				return nil, err
			}
		}
	}

	var ports nat.PortSet = nat.PortSet{}
	for _, port := range service.Ports {
		p, err := nat.NewPort("tcp", port)
		if err != nil {
			logrus.Errorf("failed to create port %s: %s", port, err)
			return nil, err
		}
		ports[p] = struct{}{}
	}

	var portMap nat.PortMap = nat.PortMap{}
	for port := range ports {
		portMap[port] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: port.Port(),
			},
		}
	}

	// create container
	resp, err := c.client.ContainerCreate(ctx, &container.Config{
		Image:        service.Name,
		ExposedPorts: ports,
		AttachStdout: true,
	}, &container.HostConfig{
		PortBindings: portMap,
	}, &network.NetworkingConfig{}, nil, service.Name)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// StartContainer starts a docker container
func (c *Docker) StartContainer(ctx context.Context, containerID string) error {
	options := types.ContainerStartOptions{}
	return c.client.ContainerStart(ctx, containerID, options)
}

// StopContainer : stop docker container
func (c *Docker) StopContainer(ctx context.Context, containerID string) error {
	timeout := 60 * time.Second
	return c.client.ContainerStop(ctx, containerID, &timeout)
}

func (d *Docker) RemoveContainer(ctx context.Context, containerID string) error {
	return d.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		Force: true,
	})
}

func (d *Docker) ContainerLogs(ctx context.Context, containerID string) error {
	logs, err := d.client.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
	})
	if err != nil {
		return err
	}

	defer logs.Close()
	jsonmessage.DisplayJSONMessagesStream(logs, os.Stdout, os.Stdout.Fd(), false, nil)
	return nil
}
