package executor

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"net/http"
	"strings"
)

type ExecuteDocker struct {
}

func (de *ExecuteDocker) StopService(service string, params string) error {
	ctx := context.Background()
	var cli http.Client
	api, err := client.NewClient("localhost", "1.37", &cli, nil)
	if err != nil {
		return err
	}
	containers, err := api.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	// BEFORE Docker stop
	listRunningDockerContainers()

	for _, container := range containers {
		if strings.Contains(container.Names[0], service) {
			// fmt.Sprintf("Stopping container {%v} with ID {%v}...", container.Names[0], container.ID[:10])
			if err := api.ContainerStop(ctx, container.ID, nil); err != nil {
				return err
			}
			// fmt.Sprintf("Successfully stopped the container for the micro-service {%v}.", service)
		}
	}
	// AFTER Docker stop
	listRunningDockerContainers()

	return nil
}

func listRunningDockerContainers() error {
	var cli http.Client
	api, err := client.NewClient("localhost", "1.37", &cli, nil)
	if err != nil {
		return err
	}

	containers, err := api.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return err
	}

	println(containers)

	return nil
}
