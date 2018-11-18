package executor

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"strings"
)

type ExecuteDocker struct {
}

func (de *ExecuteDocker) StopService(service string) error {

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.37"))
	if err != nil {
		return err
	}
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	// BEFORE Docker stop
	listRunningDockerContainers()

	for _, container := range containers {
		if strings.Contains(container.Names[0], service) {
			// fmt.Sprintf("Stopping container {%v} with ID {%v}...", container.Names[0], container.ID[:10])
			if err := cli.ContainerStop(ctx, container.ID, nil); err != nil {
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

	cli, err := client.NewClientWithOpts(client.WithVersion("1.37"))
	if err != nil {
		return err
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return err
	}

	println(containers)

	return nil
}
