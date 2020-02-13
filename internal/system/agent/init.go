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

package agent

import (
	"context"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"
	"github.com/edgexfoundry/edgex-go/internal/pkg/urlclient"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/clients"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/container"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/direct"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/executor"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/getconfig"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/setconfig"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
	"github.com/gorilla/mux"
)

// Bootstrap contains references to dependencies required by the BootstrapHandler.
type Bootstrap struct {
	router *mux.Router
}

// NewBootstrap is a factory method that returns an initialized Bootstrap receiver struct.
func NewBootstrap(router *mux.Router) *Bootstrap {
	return &Bootstrap{
		router: router,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract.  It implements agent-specific initialization.
func (b *Bootstrap) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	loadRestRoutes(b.router, dic)

	configuration := container.ConfigurationFrom(dic.Get)

	// validate metrics implementation
	switch configuration.MetricsMechanism {
	case direct.MetricsMechanism:
	case executor.MetricsMechanism:
	default:
		lc := bootstrapContainer.LoggingClientFrom(dic.Get)
		lc.Error("the requested metrics mechanism is not supported")
		return false
	}

	// add dependencies to container
	dic.Update(di.ServiceConstructorMap{
		container.GeneralClientsName: func(get di.Get) interface{} {
			return clients.NewGeneral()
		},
		container.MetricsInterfaceName: func(get di.Get) interface{} {
			logging := bootstrapContainer.LoggingClientFrom(get)
			switch configuration.MetricsMechanism {
			case direct.MetricsMechanism:
				return direct.NewMetrics(
					logging,
					container.GeneralClientsFrom(get),
					bootstrapContainer.RegistryFrom(get),
					configuration.Service.Protocol,
				)
			case executor.MetricsMechanism:
				return executor.NewMetrics(executor.CommandExecutor, logging, configuration.ExecutorPath)
			default:
				panic("unsupported metrics mechanism " + container.MetricsInterfaceName)
			}
		},
		container.OperationsInterfaceName: func(get di.Get) interface{} {
			return executor.NewOperations(
				executor.CommandExecutor,
				bootstrapContainer.LoggingClientFrom(get),
				configuration.ExecutorPath)
		},
		container.GetConfigInterfaceName: func(get di.Get) interface{} {
			logging := bootstrapContainer.LoggingClientFrom(get)
			return getconfig.New(
				getconfig.NewExecutor(
					container.GeneralClientsFrom(get),
					bootstrapContainer.RegistryFrom(get),
					logging,
					configuration.Service.Protocol),
				logging)
		},
		container.SetConfigInterfaceName: func(get di.Get) interface{} {
			return setconfig.New(setconfig.NewExecutor(bootstrapContainer.LoggingClientFrom(get), configuration))
		},
	})

	generalClients := container.GeneralClientsFrom(dic.Get)
	registryClient := bootstrapContainer.RegistryFrom(dic.Get)

	for serviceKey, serviceName := range config.ListDefaultServices() {
		generalClients.Set(
			serviceKey,
			general.NewGeneralClient(
				urlclient.New(
					registryClient != nil,
					endpoint.New(
						ctx,
						&sync.WaitGroup{},
						&registryClient,
						serviceKey,
						"/",
						internal.ClientMonitorDefault,
					),
					configuration.Clients[serviceName].Url(),
				),
			),
		)
	}

	return true
}
