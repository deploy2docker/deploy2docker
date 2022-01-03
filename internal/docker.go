/*
MIT License

Copyright (c) 2021 Deploy to Docker (satish.babariya@gmail.com)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package internal

import (
	"context"
	"os"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/go-connections/nat"
)

type Docker struct {
	client *client.Client
}

func NewDocker() (*Docker, error) {
	options := client.WithHost(LocalDockerHost)
	docker, err := client.NewClientWithOpts(options)
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

func (d *Docker) Build(ctx context.Context, path string, tags []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	if path != "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return err
		}
		dir = path
	}

	tar, err := archive.TarWithOptions(dir, &archive.TarOptions{
		ExcludePatterns: []string{".git"},
	})
	if err != nil {
		return err
	}

	defer tar.Close()

	resp, err := d.client.ImageBuild(ctx, tar, types.ImageBuildOptions{
		Tags:   tags,
		Remove: true,
	})
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	jsonmessage.DisplayJSONMessagesStream(resp.Body, os.Stdout, os.Stdout.Fd(), false, nil)

	return nil
}

func (d *Docker) IsContainerRunning(ctx context.Context, containerID string) bool {
	json, err := d.client.ContainerInspect(ctx, containerID)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false
		}
		return false
	}
	return json.State.Running
}

func (d *Docker) Run(ctx context.Context, config *Config) error {

	if d.IsContainerRunning(ctx, config.Service.Name) {
		// remove container
		err := d.client.ContainerRemove(ctx, config.Service.Name, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			return err
		}
	}

	var ports nat.PortSet = nat.PortSet{}
	for _, port := range config.Service.Ports {
		p, err := nat.NewPort("tcp", strconv.Itoa(port))
		if err != nil {
			return err
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
	resp, err := d.client.ContainerCreate(ctx, &container.Config{
		Image:        config.Service.Name,
		ExposedPorts: ports,
		AttachStdout: true,
	}, &container.HostConfig{
		PortBindings: portMap,
	}, &network.NetworkingConfig{}, nil, config.Service.Name)
	if err != nil {
		return err
	}

	// start container
	err = d.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	// get container logs
	logs, err := d.client.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
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
