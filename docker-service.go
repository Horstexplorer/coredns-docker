package coredns_docker

import (
	"context"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type DockerService struct {
	ctx context.Context
	api *client.Client
}

func NewDockerService() (*DockerService, error) {
	api, err := client.New(client.FromEnv) // todo: verify this is a good default
	if err != nil {
		return nil, err
	}

	return &DockerService{
		ctx: context.Background(),
		api: api,
	}, nil
}

func (i *DockerService) ListAllContainers() ([]container.Summary, error) {
	containers, err := i.api.ContainerList(i.ctx, client.ContainerListOptions{
		All: true,
	})

	if err != nil {
		return nil, err
	}

	return containers.Items, nil

}
