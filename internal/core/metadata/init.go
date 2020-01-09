/*******************************************************************************
 * Copyright 2017 Dell Inc.
 * Copyright (c) 2019 Intel Corporation
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
package metadata

import (
	"context"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	errorContainer "github.com/edgexfoundry/edgex-go/internal/pkg/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/notifications"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"

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

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the metadata service.
func (b *Bootstrap) BootstrapHandler(_ context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	loadRestRoutes(b.router, dic)

	configuration := container.ConfigurationFrom(dic.Get)

	// initialize clients required by service.
	registryClient := bootstrapContainer.RegistryFrom(dic.Get)

	// add dependencies to container
	dic.Update(di.ServiceConstructorMap{
		errorContainer.ErrorHandlerName: func(get di.Get) interface{} {
			return errorconcept.NewErrorHandler(bootstrapContainer.LoggingClientFrom(get))
		},
		container.CoreDataValueDescriptorClientName: func(get di.Get) interface{} {
			return coredata.NewValueDescriptorClient(
				types.EndpointParams{
					ServiceKey:  clients.CoreDataServiceKey,
					Path:        clients.ApiValueDescriptorRoute,
					UseRegistry: registryClient != nil,
					Url:         configuration.Clients["CoreData"].Url() + clients.ApiValueDescriptorRoute,
					Interval:    configuration.Service.ClientMonitor,
				},
				endpoint.Endpoint{RegistryClient: &registryClient})
		},
		container.NotificationsClientName: func(get di.Get) interface{} {
			return notifications.NewNotificationsClient(types.EndpointParams{
				ServiceKey:  clients.SupportNotificationsServiceKey,
				Path:        clients.ApiNotificationRoute,
				UseRegistry: registryClient != nil,
				Url:         configuration.Clients["Notifications"].Url() + clients.ApiNotificationRoute,
				Interval:    configuration.Service.ClientMonitor,
			},
				endpoint.Endpoint{RegistryClient: &registryClient})
		},
	})

	return true
}
