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
	"context"
	"flag"
	"sync"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/handlers"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/system/agent"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/direct"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/executor"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/getconfig"
	agentInterfaces "github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/setconfig"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/edgexfoundry/go-mod-registry/registry"
)

func httpServerBootstrapHandler(
	wg *sync.WaitGroup,
	ctx context.Context,
	startupTimer startup.Timer,
	config interfaces.Configuration,
	logging logger.LoggingClient,
	registry registry.Client) bool {

	var metricsImpl agentInterfaces.Metrics
	switch agent.Configuration.MetricsMechanism {
	case direct.MetricsMechanism:
		metricsImpl = direct.NewMetrics(logging, agent.GenClients, registry, agent.Configuration.Service.Protocol)
	case executor.MetricsMechanism:
		metricsImpl = executor.NewMetrics(executor.CommandExecutor, logging, agent.Configuration.ExecutorPath)
	default:
		logging.Error("the requested metrics mechanism is not supported")
		return false
	}

	httpServer := handlers.NewServerBootstrap(
		agent.LoadRestRoutes(
			metricsImpl,
			executor.NewOperations(executor.CommandExecutor, logging, agent.Configuration.ExecutorPath),
			getconfig.New(
				getconfig.NewExecutor(agent.GenClients, registry, logging, agent.Configuration.Service.Protocol),
				logging),
			setconfig.New(setconfig.NewExecutor(logging, agent.Configuration))))
	return httpServer.Handler(wg, ctx, startupTimer, config, logging, registry)
}

func main() {
	startupTimer := startup.NewStartUpTimer(1, internal.BootTimeoutDefault)

	var useRegistry bool
	var configDir, profileDir string

	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use registry service.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use registry service.")
	flag.StringVar(&profileDir, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&profileDir, "p", "", "Specify a profile other than default.")
	flag.StringVar(&configDir, "confdir", "", "Specify local configuration directory")

	flag.Usage = usage.HelpCallback
	flag.Parse()

	bootstrap.Run(
		configDir,
		profileDir,
		internal.ConfigFileName,
		useRegistry,
		clients.SystemManagementAgentServiceKey,
		agent.Configuration,
		startupTimer,
		[]interfaces.BootstrapHandler{
			handlers.SecretClientBootstrapHandler,
			agent.BootstrapHandler,
			httpServerBootstrapHandler,
			handlers.NewStartMessage(clients.SystemManagementAgentServiceKey, edgex.Version).Handler,
		})
}
