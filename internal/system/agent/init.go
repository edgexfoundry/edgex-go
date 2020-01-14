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
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/delegate"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/common/middleware/debugging"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/common"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/correlationid"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/router"
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
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"

	"github.com/gorilla/mux"
)

// Bootstrap contains references to dependencies required by the BootstrapHandler.
type Bootstrap struct {
	muxRouter              *mux.Router
	inDebugMode            bool
	inV2AcceptanceTestMode bool
}

// NewBootstrap is a factory method that returns an initialized Bootstrap receiver struct.  inDebug determines
// whether or not the debug middleware is installed.  inV2AcceptanceTestMode determines if the service is running in
// the test runner context (in which case, we shouldn't load the APIv1 routes).
func NewBootstrap(muxRouter *mux.Router, inDebugMode, inV2AcceptanceTestMode bool) *Bootstrap {
	return &Bootstrap{
		muxRouter:              muxRouter,
		inDebugMode:            inDebugMode,
		inV2AcceptanceTestMode: inV2AcceptanceTestMode,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract.  It implements agent-specific initialization.
func (b *Bootstrap) BootstrapHandler(_ context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)

	if !b.inV2AcceptanceTestMode {
		loadV1Routes(b.muxRouter, dic)
	}
	b.loadV2Routes(dic, lc)

	// validate metrics implementation
	switch configuration.MetricsMechanism {
	case direct.MetricsMechanism:
	case executor.MetricsMechanism:
	default:
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

	// initialize clients required by service.
	generalClients := container.GeneralClientsFrom(dic.Get)
	registryClient := bootstrapContainer.RegistryFrom(dic.Get)
	for serviceKey, serviceName := range config.ListDefaultServices() {
		generalClients.Set(
			serviceKey,
			general.NewGeneralClient(
				types.EndpointParams{
					ServiceKey:  serviceKey,
					Path:        "/",
					UseRegistry: registryClient != nil,
					Url:         configuration.Clients[serviceName].Url(),
					Interval:    internal.ClientMonitorDefault,
				},
				endpoint.Endpoint{RegistryClient: &registryClient}))
	}

	return true
}

// loadV2Routes creates a new command-query router and handles the related mux.Router initialization for API V2 routes.
func (b *Bootstrap) loadV2Routes(_ *di.Container, lc logger.LoggingClient) {
	correlationid.WireUp(b.muxRouter)

	handlers := []delegate.Handler{}
	if b.inDebugMode {
		handlers = append(handlers, debugging.New(lc).Handler)
	}

	router.Initialize(
		b.muxRouter,
		handlers,
		common.V2Routes(
			b.inV2AcceptanceTestMode,
			[]router.Controller{},
		),
	)
}
