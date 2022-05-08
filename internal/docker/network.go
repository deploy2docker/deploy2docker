package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
)

// CreateBridgeNetwork creates a docker bridge network
func (c *Docker) CreateBridgeNetwork(ctx context.Context, name string) (types.NetworkCreateResponse, error) {
	n, _ := c.GetNetworkByName(name)

	if n.ID != "" {
		return types.NetworkCreateResponse{ID: n.ID}, nil
	}

	return c.client.NetworkCreate(ctx, name, types.NetworkCreate{
		Driver:         "bridge",
		CheckDuplicate: true,
	})
}

// GetNetworkByName returns the network (if it exist) from the docker host
func (c *Docker) GetNetworkByName(name string) (types.NetworkResource, error) {
	ctx := context.Background()
	defer ctx.Done()

	networks, err := c.client.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return types.NetworkResource{}, err
	}

	for _, net := range networks {
		if name == net.Name {
			return net, nil
		}
	}

	return types.NetworkResource{}, fmt.Errorf("network %s not found", name)
}
