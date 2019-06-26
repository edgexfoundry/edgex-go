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
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
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
	METRICS                     = "metrics"
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

		LoggingClient.Debug(fmt.Sprintf("service: %s", service))
		LoggingClient.Debug(fmt.Sprintf("operation: %s", operation))

		// Don't run commands for unknown services - could be an attack
		if agent.IsKnownServiceKey(service) {

			response, err := executeDockerCommands(service, operation)
			if err != nil {
				LoggingClient.Error(fmt.Sprintf("error performing  %s on service %s: %v", operation, service, err.Error()))
			} else {
				LoggingClient.Info(fmt.Sprintf("success performing %s on service %s", operation, service))
				LoggingClient.Debug(fmt.Sprintf("response from main: %s", response))
				LoggingClient.Debug(fmt.Sprintf("the IsKnownServiceKey() check was a success for service %s", service))
				LoggingClient.Debug(fmt.Sprintf("operation: %s", operation))
			}
		} else {
			LoggingClient.Error(fmt.Sprintf("the service %s is an unknown one for which the request was made to run Docker command", service))
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

func executeDockerCommands(service string, operation string) (string, error) {

	// Validate that a known operation was requested.
	if operation == START || operation == STOP || operation == RESTART {
		// run the docker command
		out, err := exec.Command("docker", operation, service).CombinedOutput()
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("docker command failed: %s", string(out)))
			return string(out), err
		}

		// Check that the command actually resulted in the correct state for the container
		var expectedStatus bool
		switch operation {
		case START:
			fallthrough
		case RESTART:
			expectedStatus = true
		case STOP:
			expectedStatus = false
		case METRICS:

		default:
			panic(fmt.Sprintf("invalid operation %s", operation))
		}

		correctStatus, err := checkDockerContainerStatus(service, expectedStatus)
		if err != nil {
			return "", err
		}
		if !correctStatus {
			return "", fmt.Errorf("docker %s operation failed for service %s", operation, service)
		}
		return "", nil
	} else if operation == METRICS {

		LoggingClient.Info(fmt.Sprintf("run the docker stats on service (%s) for operation (%s)", service, operation))
		cmd := exec.Command("docker", "stats")
		LoggingClient.Info(fmt.Sprintf("finished running the docker stats command"))

		var stdout, stderr []byte
		var errStdout, errStderr error
		stdoutIn, _ := cmd.StdoutPipe()
		stderrIn, _ := cmd.StderrPipe()
		err := cmd.Start()
		if err != nil {
			LoggingClient.Info(fmt.Sprintf("cmd.Start() failed with '%s'\n", err.Error()))
		}

		// Note that cmd.Wait() should be called only AFTER we finish reading
		// from stdoutIn and stderrIn.
		// The wait group (wg) ensures that we finish
		var wg sync.WaitGroup
		wg.Add(1)
		ticker := time.NewTicker(time.Second)
		go func(ticker *time.Ticker) {
			stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn)
			wg.Done()
			now := time.Now()
			for range ticker.C {
				LoggingClient.Info(fmt.Sprintf("%s", time.Since(now)))
			}
		}(ticker)

		// Create a timer that will kill the process
		timer := time.NewTimer(time.Second * 3)
		go func(timer *time.Timer, ticker *time.Ticker, cmd *exec.Cmd) {
			for range timer.C {
				cmd.Process.Signal(os.Kill)
				ticker.Stop()
			}
		}(timer, ticker, cmd)
		stderr, errStderr = copyAndCapture(os.Stderr, stderrIn)

		wg.Wait()

		err = cmd.Wait()
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("cmd.Run() failed with %s\n", err.Error()))
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
		if errStdout != nil || errStderr != nil {
			// The reason for using log.Fatal() here is that we are intentionally terminating the main program.
			// The intent is not to recover gracefully from the error which has been logged already through
			// the LoggingClient.
			LoggingClient.Error(fmt.Sprintf("failed to capture stdout or stderr\n"))
			log.Fatal("failed to capture stdout or stderr\n")
		}
		outStr, errStr := string(stdout), string(stderr)
		LoggingClient.Debug(fmt.Sprintf("\nout:\n%s\nerr:\n%s\n", outStr, errStr))
		LoggingClient.Info("invocation of metrics executor in runDockerCommands() succeeded...")
		return outStr, nil
	} else {
		return "", fmt.Errorf("system management was requested to perform an unknown operation %s on the service %s", operation, service)
	}
}

func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}
