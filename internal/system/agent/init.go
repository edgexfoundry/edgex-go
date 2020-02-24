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

	"github.com/edgexfoundry/edgex-go/internal/system/agent/container"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/direct"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/executor"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/factory"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/getconfig"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/setconfig"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

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

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)

	// validate metrics implementation
	switch configuration.MetricsMechanism {
	case direct.MetricsMechanism:
	case executor.MetricsMechanism:
	default:
		lc.Error("the requested metrics mechanism is not supported")
		return false
	}

	clientFactory := factory.New(
		ctx,
		wg,
		bootstrapContainer.RegistryFrom(dic.Get),
		configuration.Clients,
		configuration.Service.Protocol,
	)

	// add dependencies to container
	dic.Update(di.ServiceConstructorMap{
		container.MetricsInterfaceName: func(get di.Get) interface{} {
			switch configuration.MetricsMechanism {
			case direct.MetricsMechanism:
				return direct.NewMetrics(clientFactory)
			case executor.MetricsMechanism:
				return executor.NewMetrics(executor.CommandExecutor, lc, configuration.ExecutorPath)
			default:
				panic("unsupported metrics mechanism " + container.MetricsInterfaceName)
			}
		},
		container.OperationsInterfaceName: func(get di.Get) interface{} {
			return executor.NewOperations(executor.CommandExecutor, lc, configuration.ExecutorPath)
		},
		container.GetConfigInterfaceName: func(get di.Get) interface{} {
			return getconfig.New(getconfig.NewExecutor(clientFactory), lc)
		},
		container.SetConfigInterfaceName: func(get di.Get) interface{} {
			return setconfig.New(setconfig.NewExecutor(lc, configuration))
		},
	})

	return true
}
