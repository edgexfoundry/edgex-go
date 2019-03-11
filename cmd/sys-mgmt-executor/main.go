/*******************************************************************************
 * Copyright 2019 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/system/agent"
)

// Global variables
var err error

var usageStr = `
Usage: ./main service operation		Start app with requested {service} and {operation}
       -h							Show this message
`

const (
	START      = "start"
	STOP       = "stop"
	RESTART    = "restart"
	AppOpenMsg = "This is the docker-compose-executor application!"
)

// For use explicitly by this SMA executor, and given that Docker recognizes these
// labels (e.g. "Notifications" and not "edgex-support-notifications", and likewise)
// for other services, this simple remapping takes care of that wherby exec.Command
// can work. Hence the following map definition.
var services = map[string]string{
	internal.SupportNotificationsServiceKey: "Notifications",
	internal.CoreCommandServiceKey:          "Command",
	internal.CoreDataServiceKey:             "CoreData",
	internal.CoreMetaDataServiceKey:         "Metadata",
	internal.ExportClientServiceKey:         "Export",
	internal.ExportDistroServiceKey:         "Distro",
	internal.SupportLoggingServiceKey:       "Logging",
	internal.SupportSchedulerServiceKey:     "Scheduler",
}

// usage will print out the flag options for the app.
// This function is based on usage.go (in internal / pkg / usage)
func HelpCallback() {
	msg := fmt.Sprintf(usageStr, os.Args[0])
	fmt.Printf("%s\n", msg)
	os.Exit(0)
}

func main() {

	start := time.Now()

	flag.Usage = HelpCallback
	flag.Parse()

	agent.LoggingClient.Info(AppOpenMsg)

	// Time it took to start service
	agent.LoggingClient.Info("Application started in: " + time.Since(start).String())

	var service = ""
	var operation = ""

	if len(os.Args) > 2 {
		service = os.Args[1]
		operation = os.Args[2]

		err := ExecuteDockerCommands(service, operation)

		if err != nil {
			agent.LoggingClient.Error(fmt.Sprintf("error performing  %s on service %s: %v", operation, service, err.Error()))
		} else {
			agent.LoggingClient.Info(fmt.Sprintf("success performing %s on service %s", operation, service))
		}
	}
}

func findDockerContainerStatus(service string, status string) bool {

	var (
		cmdOut []byte
		err    error
	)
	cmdName := "docker"
	cmdArgs := []string{"ps"}
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).CombinedOutput(); err != nil {
		agent.LoggingClient.Error(err.Error())
		os.Exit(1)
	}

	dockerOutput := string(cmdOut)

	// Find whether the container to start has started.
	for _, line := range strings.Split(strings.TrimSuffix(dockerOutput, "\n"), "\n") {
		if strings.Contains(line, service) {

			if status == "Up" {
				if strings.Contains(line, "Up") {
					agent.LoggingClient.Info(fmt.Sprintf("container for service %s started: %s", service, line))
					return true
				} else {
					agent.LoggingClient.Warn(fmt.Sprintf("container for service %s NOT started", service))
					return false
				}
			} else if status == "Exited" {
				if strings.Contains(line, "Exited") {
					agent.LoggingClient.Info(fmt.Sprintf("container for service %s stopped: %s", service, line))
					return true
				} else {
					agent.LoggingClient.Warn(fmt.Sprintf("container for service %s NOT stopped", service))
					return false
				}
			}
		}
	}
	return false
}

func ExecuteDockerCommands(service string, operation string) error {
	_, knownService := services[service]

	if knownService {
		err := runDockerCommands(service, services[service], operation)
		agent.LoggingClient.Error(fmt.Sprintf("service %s ran into error while running Docker command: %v", service, err))
		return err
	} else {
		err := errors.New("is an unknown service for which request was made to run Docker command")
		agent.LoggingClient.Error(fmt.Sprintf("the service %s: %v", service, err))
		return err
	}
}

func runDockerCommands(service string, dockerService string, operation string) error {

	var (
		err    error
		cmdDir string
	)

	cmdName := "docker"

	cmdArgs := []string{operation, dockerService}

	// Validate that a known operation was requested.
	if operation == START || operation == STOP || operation == RESTART {

		cmd := exec.Command(cmdName, cmdArgs...)
		cmd.Dir = cmdDir

		out, err := cmd.CombinedOutput()
		if err != nil {
			agent.LoggingClient.Error(err.Error())
			agent.LoggingClient.Info("docker command failed: %s", string(out))
		}

		switch operation {
		case START:
		case RESTART:
			if !findDockerContainerStatus(service, "Up") {
				agent.LoggingClient.Warn("docker start operation failed for service %s", service)
			}
			break

		case STOP:
			if !findDockerContainerStatus(service, "Exited") {
				agent.LoggingClient.Warn("docker stop operation failed for service %s", service)
			}
			break

		default:
			agent.LoggingClient.Warn("unknown operation %s was requested", operation)
			break
		}
	} else {
		err := errors.New("an unknown operation")
		agent.LoggingClient.Error(fmt.Sprintf("system management was requested to perform %v %s on the %s service", err.Error(), operation, service))
		return err
	}

	return err
}
