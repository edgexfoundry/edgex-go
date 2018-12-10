package executor

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/logger"
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
			logs.LoggingClient.Debug("stopping container", "container name", container.Names[0], "container id", container.ID[:10])
			if err := cli.ContainerStop(ctx, container.ID, nil); err != nil {
				return err
			}
			logs.LoggingClient.Debug("successfully stopped container", "service name", service)
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
