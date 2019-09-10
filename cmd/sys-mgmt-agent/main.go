/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
	"os/signal"
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/system/agent"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/direct"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/executor"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func main() {
	start := time.Now()
	var useRegistry bool
	var configDir, profileDir string

	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use registry service.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use registry service.")
	flag.StringVar(&profileDir, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&profileDir, "p", "", "Specify a profile other than default.")
	flag.StringVar(&configDir, "confdir", "", "Specify local configuration directory")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	instance := agent.NewInstance()
	params := startup.BootParams{
		UseRegistry: useRegistry,
		ConfigDir:   configDir,
		ProfileDir:  profileDir,
		BootTimeout: internal.BootTimeoutDefault,
	}
	startup.Bootstrap(params, instance.Retry, logBeforeInit)

	ok := instance.Init(useRegistry)
	if !ok {
		logBeforeInit(fmt.Errorf("%s: service bootstrap failed", clients.SystemManagementAgentServiceKey))
		os.Exit(1)
	}

	instance.LoggingClient.Info("Service dependencies resolved...")
	instance.LoggingClient.Info(fmt.Sprintf("Starting %s %s ", clients.SystemManagementAgentServiceKey, edgex.Version))

	instance.LoggingClient.Info(instance.Configuration.Service.StartupMsg)

	errs := make(chan error, 2)
	listenForInterrupt(errs)
	startup.StartHTTPServer(
		instance.LoggingClient,
		instance.Configuration.Service.Timeout,
		agent.LoadRestRoutes(instance, getMetricsImplementation(instance)),
		instance.Configuration.Service.Host+":"+strconv.Itoa(instance.Configuration.Service.Port),
		errs)

	// Time it took to start service
	instance.LoggingClient.Info("Service started in: " + time.Since(start).String())
	instance.LoggingClient.Info("Listening on port: " + strconv.Itoa(instance.Configuration.Service.Port))
	c := <-errs
	instance.Destruct()
	instance.LoggingClient.Warn(fmt.Sprintf("terminating: %v", c))

	os.Exit(0)
}

// getMetricsImplementation creates and returns an interfaces.Metrics implementation based on configuration settings.
func getMetricsImplementation(instance *agent.Instance) interfaces.Metrics {
	var result interfaces.Metrics
	switch instance.Configuration.MetricsMechanism {
	case direct.MetricsMechanism:
		result = direct.NewMetrics(
			instance.LoggingClient,
			instance.GenClients,
			instance.Configuration.Clients,
			instance.RegistryClient,
			instance.Configuration.Service.Protocol)
	case executor.MetricsMechanism:
		result = executor.NewMetrics(
			executor.CommandExecutor,
			instance.LoggingClient,
			instance.Configuration.ExecutorPath)
	default:
		instance.LoggingClient.Error("the requested metrics mechanism is not supported")
		os.Exit(1)
	}
	return result
}

func logBeforeInit(err error) {
	l := logger.NewClient(clients.SystemManagementAgentServiceKey, false, "", models.InfoLog)
	l.Error(err.Error())
}

func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}
