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
	"encoding/json"
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
var usageStr = `Usage: ./%s <service> <operation>		Start app with requested {service} and {operation}
       -h							Show this message`

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
	fmt.Println(msg)
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

func checkDockerContainerStatus(service string, running bool) (bool, error) {
	// check the status of the container using the json format - include all
	// containers as the container we want to check may be Exited
	cmdOut, err := exec.Command("docker", "inspect", service).CombinedOutput()
	if err != nil {
		LoggingClient.Error(err.Error())
		os.Exit(1)
	}

	dec := json.NewDecoder(strings.NewReader(string(cmdOut)))
	type containerInfo struct {
		State struct {
			Running bool
		}
	}

	c := []containerInfo{}
	for {
		err = dec.Decode(&c)
		if err != nil {
			return false, err
		}
		switch {
		case len(c) < 1:
			return false, fmt.Errorf("container %s not found", service)
		case len(c) > 1:
			return false, fmt.Errorf("multiple containers found with name %s", service)
		default:
			if c[0].State.Running {
				LoggingClient.Info(fmt.Sprintf("service container %s is running", service))
			} else {
				LoggingClient.Info(fmt.Sprintf("service container %s is not running", service))
			}
			return c[0].State.Running == running, err
		}
	}
}

func ExecuteDockerCommands(service string, operation string) error {
	// don't run commands for unknown services - could be an attack
	if agent.IsKnownServiceKey(service) {
		return runDockerCommands(service, operation)
	}
	return fmt.Errorf("the service %s is an unknown service for which request was made to run Docker command", service)
}

func runDockerCommands(service string, operation string) error {
	// Validate that a known operation was requested.
	if operation == START || operation == STOP || operation == RESTART {
		// run the docker command
		out, err := exec.Command("docker", operation, service).CombinedOutput()
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("docker command failed: %s", string(out)))
			return err
		}

		// check that the command actually resulted in the correct state for
		// the container
		var expectedStatus bool
		switch operation {
		case START:
			fallthrough
		case RESTART:
			expectedStatus = true
		case STOP:
			expectedStatus = false
		default:
			// impossible to get here
			panic(fmt.Sprintf("invalid operation %s", operation))
		}

		correctStatus, err := checkDockerContainerStatus(service, expectedStatus)
		if err != nil {
			return err
		}
		if !correctStatus {
			return fmt.Errorf("docker %s operation failed for service %s", operation, service)
		}
		return nil
	}
	return fmt.Errorf("system management was requested to perform an unknown operation %s on the service %s", operation, service)
}
