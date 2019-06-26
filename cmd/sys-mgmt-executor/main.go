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
	"time"

	"github.com/edgexfoundry/edgex-go/internal/system/agent"
	"github.com/edgexfoundry/edgex-go/internal/system/executor"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// Global variables
var usageStr = `Usage: ./%s <service> <operation>		Start app with requested {service} and {operation}
       -h							Show this message`

const (
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
	LoggingClient := logger.NewClient(SystemManagementExecutorKey, EnableRemote, LoggingTarget, LogLevel)

	LoggingClient.Info(AppOpenMsg)

	// Time it took to start service
	LoggingClient.Info("Application started in: " + time.Since(start).String())

	if len(os.Args) > 2 {
		service := os.Args[1]
		operation := os.Args[2]

		// Don't run commands for unknown services - could be an attack
		if agent.IsKnownServiceKey(service) {
			result, err := executor.Execute(operation, service, func(arg ...string) ([]byte, error) {
				return exec.Command("docker", arg...).CombinedOutput()
			})
			if err != nil {
				LoggingClient.Error(fmt.Sprintf("error performing %s on service %s: %v", operation, service, err.Error()))
				os.Exit(1)
			} else {
				LoggingClient.Info(fmt.Sprintf("success performing %s on service %s", operation, service))
				LoggingClient.Debug(fmt.Sprintf("the IsKnownServiceKey() check was a success for service %s", service))
				LoggingClient.Debug(fmt.Sprintf("operation: %s", operation))

				if result != nil {
					// Push result to stdout for consumption by system management agent
					fmt.Printf("%s", string(result))
					os.Exit(0)
				}
			}
		} else {
			LoggingClient.Error(fmt.Sprintf("the service %s is an unknown one for which the request was made to run Docker command", service))
			os.Exit(1)
		}
	}
}
