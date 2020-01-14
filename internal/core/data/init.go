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
package data

import (
	"context"
	"fmt"
	"sync"

	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	errorContainer "github.com/edgexfoundry/edgex-go/internal/pkg/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/delegate"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/common/middleware/debugging"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/common"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/correlationid"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/router"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"

	"github.com/edgexfoundry/go-mod-messaging/messaging"
	msgTypes "github.com/edgexfoundry/go-mod-messaging/pkg/types"

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

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the data service.
func (b *Bootstrap) BootstrapHandler(_ context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	configuration := dataContainer.ConfigurationFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	if !b.inV2AcceptanceTestMode {
		loadV1Routes(b.muxRouter, dic)
	}
	b.loadV2Routes(dic, lc)

	// initialize clients required by service.
	registryClient := container.RegistryFrom(dic.Get)
	mdc := metadata.NewDeviceClient(
		types.EndpointParams{
			ServiceKey:  clients.CoreMetaDataServiceKey,
			Path:        clients.ApiDeviceRoute,
			UseRegistry: registryClient != nil,
			Url:         configuration.Clients["Metadata"].Url() + clients.ApiDeviceRoute,
			Interval:    configuration.Service.ClientMonitor,
		},
		endpoint.Endpoint{RegistryClient: &registryClient})
	msc := metadata.NewDeviceServiceClient(
		types.EndpointParams{
			ServiceKey:  clients.CoreMetaDataServiceKey,
			Path:        clients.ApiDeviceServiceRoute,
			UseRegistry: registryClient != nil,
			Url:         configuration.Clients["Metadata"].Url() + clients.ApiDeviceRoute,
			Interval:    configuration.Service.ClientMonitor,
		},
		endpoint.Endpoint{RegistryClient: &registryClient})

	// Create the messaging client
	msgClient, err := messaging.NewMessageClient(
		msgTypes.MessageBusConfig{
			PublishHost: msgTypes.HostInfo{
				Host:     configuration.MessageQueue.Host,
				Port:     configuration.MessageQueue.Port,
				Protocol: configuration.MessageQueue.Protocol,
			},
			Type:     configuration.MessageQueue.Type,
			Optional: configuration.MessageQueue.Optional,
		})

	if err != nil {
		lc.Error(fmt.Sprintf("failed to create messaging client: %s", err.Error()))
		return false
	}

	chEvents := make(chan interface{}, 100)
	// initialize event handlers
	initEventHandlers(lc, chEvents, mdc, msc, configuration)

	dic.Update(di.ServiceConstructorMap{
		dataContainer.MetadataDeviceClientName: func(get di.Get) interface{} {
			return mdc
		},
		dataContainer.MessagingClientName: func(get di.Get) interface{} {
			return msgClient
		},
		dataContainer.EventsChannelName: func(get di.Get) interface{} {
			return chEvents
		},
		errorContainer.ErrorHandlerName: func(get di.Get) interface{} {
			return errorconcept.NewErrorHandler(lc)
		},
	})

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
