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
	v2 "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container/v2"
	errorContainer "github.com/edgexfoundry/edgex-go/internal/pkg/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/delegate"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/common/middleware/debugging"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/common"
	addressableCreate "github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/core/metadata/addressable/create"
	addressableDelete "github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/core/metadata/addressable/delete"
	addressableRead "github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/core/metadata/addressable/read"
	addressableUpdate "github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/core/metadata/addressable/update"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/correlationid"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/router"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/notifications"
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

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the metadata service.
func (b *Bootstrap) BootstrapHandler(_ context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	configuration := container.ConfigurationFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	registryClient := bootstrapContainer.RegistryFrom(dic.Get)

	if !b.inV2AcceptanceTestMode {
		loadV1Routes(b.muxRouter, dic)
	}
	b.loadV2Routes(dic, lc)

	// add dependencies to container
	dic.Update(di.ServiceConstructorMap{
		errorContainer.ErrorHandlerName: func(get di.Get) interface{} {
			return errorconcept.NewErrorHandler(lc)
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

// loadV2Routes creates a new command-query router and handles the related mux.Router initialization for API V2 routes.
func (b *Bootstrap) loadV2Routes(dic *di.Container, lc logger.LoggingClient) {
	correlationid.WireUp(b.muxRouter)

	handlers := []delegate.Handler{}
	if b.inDebugMode {
		handlers = append(handlers, debugging.New(lc).Handler)
	}

	persistence := v2.PersistenceFrom(dic.Get)
	service := persistence.Metadata()
	addressable := service.Addressable()

	router.Initialize(
		b.muxRouter,
		handlers,
		common.V2Routes(
			b.inV2AcceptanceTestMode,
			[]router.Controller{
				addressableCreate.New(addressable),
				addressableUpdate.New(addressable),
				addressableRead.New(addressable),
				addressableDelete.New(addressable),
			},
		),
	)
}
