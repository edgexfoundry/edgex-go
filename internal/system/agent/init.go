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

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/system/agent/application"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/application/direct"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/application/executor"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/container"
)

// Bootstrap contains references to dependencies required by the BootstrapHandler.
type Bootstrap struct {
	router      *mux.Router
	serviceName string
}

// NewBootstrap is a factory method that returns an initialized Bootstrap receiver struct.
func NewBootstrap(router *mux.Router, serviceName string) *Bootstrap {
	return &Bootstrap{
		router:      router,
		serviceName: serviceName,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract.  It implements agent-specific initialization.
func (b *Bootstrap) BootstrapHandler(_ context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	LoadRestRoutes(b.router, dic, b.serviceName)

	configuration := container.ConfigurationFrom(dic.Get)

	// validate metrics implementation
	switch configuration.MetricsMechanism {
	case application.DirectMechanism:
	case application.ExecutorMechanism:
	default:
		lc := bootstrapContainer.LoggingClientFrom(dic.Get)
		lc.Error("the requested metrics mechanism is not supported")
		return false
	}

	// add dependencies to container
	dic.Update(di.ServiceConstructorMap{
		container.V2MetricsInterfaceName: func(get di.Get) interface{} {
			lc := bootstrapContainer.LoggingClientFrom(get)
			switch configuration.MetricsMechanism {
			case application.DirectMechanism:
				rc := bootstrapContainer.RegistryFrom(get)
				return direct.NewMetrics(lc, rc, configuration)
			case application.ExecutorMechanism:
				return executor.NewMetrics(executor.CommandExecutor, lc, configuration.ExecutorPath)
			default:
				panic("unsupported metrics mechanism " + configuration.MetricsMechanism)
			}
		},
	})

	return true
}
