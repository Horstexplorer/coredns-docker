package docker

import (
	"context"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type DockerApiUtil struct {
	ctx context.Context
	api *client.Client
}

func NewDockerService() (*DockerApiUtil, error) {

	Logger.Debug("Initializing Docker API client...")

	api, err := client.New(client.FromEnv)
	if err != nil {
		return nil, err
	}

	return &DockerApiUtil{
		ctx: context.Background(),
		api: api,
	}, nil
}

func (i *DockerApiUtil) ListAllContainers() ([]container.Summary, error) {
	containers, err := i.api.ContainerList(i.ctx, client.ContainerListOptions{
		All: true,
	})

	if err != nil {
		return nil, err
	}

	return containers.Items, nil

}
