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
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/system/agent"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// Global variables
var LoggingClient logger.LoggingClient
var usageStr = `
Usage: ./sys-mgmt-executor <service> <operation>		Start app with requested {service} and {operation}
       -h							Show this message
`

const (
	START                       = "start"
	STOP                        = "stop"
	RESTART                     = "restart"
	SystemManagementExecutorKey = "docker-compose-executor"
	AppOpenMsg                  = "This is the docker-compose-executor application!"
	LoggingTarget               = "console"
	EnableRemote                = false
	LogLevel                    = "INFO"
)

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

	// Setup Logging
	LoggingClient = logger.NewClient(SystemManagementExecutorKey, EnableRemote, LoggingTarget, LogLevel)

	LoggingClient.Info(AppOpenMsg)

	// Time it took to start service
	LoggingClient.Info("Application started in: " + time.Since(start).String())

	var service = ""
	var operation = ""

	if len(os.Args) > 2 {
		service = os.Args[1]
		operation = os.Args[2]

		err := ExecuteDockerCommands(service, operation)

		if err != nil {
			LoggingClient.Error(fmt.Sprintf("error performing  %s on service %s: %v", operation, service, err.Error()))
		} else {
			LoggingClient.Info(fmt.Sprintf("success performing %s on service %s", operation, service))
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
		LoggingClient.Error(err.Error())
		os.Exit(1)
	}

	dockerOutput := string(cmdOut)

	// Find whether the container to start has started.
	for _, line := range strings.Split(strings.TrimSuffix(dockerOutput, "\n"), "\n") {
		if strings.Contains(line, service) {

			if status == "Up" {
				if strings.Contains(line, "Up") {
					LoggingClient.Info(fmt.Sprintf("container for service %s started: %s", service, line))
					return true
				} else {
					LoggingClient.Warn(fmt.Sprintf("container for service %s NOT started", service))
					return false
				}
			} else if status == "Exited" {
				if strings.Contains(line, "Exited") {
					LoggingClient.Info(fmt.Sprintf("container for service %s stopped: %s", service, line))
					return true
				} else {
					LoggingClient.Warn(fmt.Sprintf("container for service %s NOT stopped", service))
					return false
				}
			}
		}
	}
	return false
}

func ExecuteDockerCommands(service string, operation string) error {

	if agent.IsKnownServiceKey(service) {
		err := runDockerCommands(service, operation)
		LoggingClient.Error(fmt.Sprintf("service %s ran into error while running Docker command: %v", service, err))
		return err
	} else {
		err := fmt.Errorf("the service %s is an unknown service for which request was made to run Docker command", service)
		LoggingClient.Error(err.Error())
		return err
	}
}

func runDockerCommands(service string, operation string) error {

	var (
		err    error
		cmdDir string
	)

	cmdName := "docker"

	cmdArgs := []string{operation, service}

	// Validate that a known operation was requested.
	if operation == START || operation == STOP || operation == RESTART {

		cmd := exec.Command(cmdName, cmdArgs...)
		cmd.Dir = cmdDir

		out, err := cmd.CombinedOutput()
		if err != nil {
			LoggingClient.Error(err.Error())
			LoggingClient.Info("docker command failed: %s", string(out))
		}

		switch operation {
		case START:
		case RESTART:
			if !findDockerContainerStatus(service, "Up") {
				LoggingClient.Warn("docker start operation failed for service %s", service)
			}
			break

		case STOP:
			if !findDockerContainerStatus(service, "Exited") {
				LoggingClient.Warn("docker stop operation failed for service %s", service)
			}
			break

		default:
			LoggingClient.Warn("unknown operation %s was requested", operation)
			break
		}
	} else {
		err := fmt.Errorf("system management was requested to perform an unknown operation %s on the service %s", operation, service)
		LoggingClient.Error(err.Error())
		return err

	}

	return err
}
