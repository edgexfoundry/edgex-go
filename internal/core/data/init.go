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

	"github.com/edgexfoundry/go-mod-core-contracts/clients/urlclient/local"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/edgex-go/internal/core/data/v2"
	v2DataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/v2/bootstrap/container"
	errorContainer "github.com/edgexfoundry/edgex-go/internal/pkg/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/error"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	"github.com/edgexfoundry/go-mod-messaging/messaging"
	msgTypes "github.com/edgexfoundry/go-mod-messaging/pkg/types"

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

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the data service.
func (b *Bootstrap) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	loadRestRoutes(b.router, dic)
	v2.LoadRestRoutes(b.router, dic)

	configuration := dataContainer.ConfigurationFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	mdc := metadata.NewDeviceClient(local.New(configuration.Clients["Metadata"].Url() + clients.ApiDeviceRoute))
	msc := metadata.NewDeviceServiceClient(local.New(configuration.Clients["Metadata"].Url() + clients.ApiDeviceRoute))

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

	for startupTimer.HasNotElapsed() {
		err = msgClient.Connect()
		if err == nil {
			break
		}

		lc.Warn(fmt.Sprintf("couldn't connect to message bus: %s", err.Error()))
		startupTimer.SleepForInterval()
	}

	if err != nil {
		lc.Error(fmt.Sprintf("failed to connect to message bus in allotted time"))
		return false
	}

	// Setup special "defer" go func that will disconnect from the message bus when the service is exiting
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				if err := msgClient.Disconnect(); err != nil {
					lc.Error("failed to disconnect from the Message Bus")
					return
				}
				lc.Info("Message Bus disconnected")
				return
			}
		}
	}()

	lc.Info(fmt.Sprintf(
		"Connected to %s Message Bus @ %s://%s:%d publishing on '%s' topic",
		configuration.MessageQueue.Type,
		configuration.MessageQueue.Protocol,
		configuration.MessageQueue.Host,
		configuration.MessageQueue.Port,
		configuration.MessageQueue.Topic))

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
		v2DataContainer.MetadataDeviceClientName: func(get di.Get) interface{} { // add v2 API MetadataDeviceClient
			return mdc
		},
		v2DataContainer.ErrorHandlerName: func(get di.Get) interface{} { // add v2 API error handler
			return error.NewErrorHandler(lc)
		},
	})

	return true
}
