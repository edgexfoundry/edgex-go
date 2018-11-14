package executor

import (
	"context"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// ExecuteDocker implements ServiceStopper and ServiceStarter to
// start and stop services that are running as docker containers
type ExecuteDocker struct {
	ComposeURL string
}

// Stop stops a service using the docker cli directly
func (de *ExecuteDocker) Stop(service string, params []string) error {

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

// Start starts a service using docker compose
func (de *ExecuteDocker) Start(service string, params []string) error {
	return StartDockerContainerCompose(service, de.ComposeURL)
}

// Restart restarts the service using the docker cli and docker compose
func (de *ExecuteDocker) Restart(service string, params []string) error {
	err := de.Stop(service, params)
	if err != nil {
		return err
	}
	// the time to sleep should probably be a paramater passed into here
	time.Sleep(time.Second * time.Duration(1))
	return de.Start(service, params)
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
