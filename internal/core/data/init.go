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
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"

	"github.com/edgexfoundry/go-mod-messaging/messaging"
	msgTypes "github.com/edgexfoundry/go-mod-messaging/pkg/types"
)

// Global variables
var Configuration = &ConfigurationStruct{}

// TODO: Refactor names in separate PR: See comments on PR #1133
var msc metadata.DeviceServiceClient

var httpErrorHandler errorconcept.ErrorHandler

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the data service.
func BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	// update global variables.
	loggingClient := container.LoggingClientFrom(dic.Get)
	httpErrorHandler = errorconcept.NewErrorHandler(loggingClient)

	// initialize clients required by service.
	registryClient := container.RegistryFrom(dic.Get)
	mdc := metadata.NewDeviceClient(
		types.EndpointParams{
			ServiceKey:  clients.CoreMetaDataServiceKey,
			Path:        clients.ApiDeviceRoute,
			UseRegistry: registryClient != nil,
			Url:         Configuration.Clients["Metadata"].Url() + clients.ApiDeviceRoute,
			Interval:    Configuration.Service.ClientMonitor,
		},
		endpoint.Endpoint{RegistryClient: &registryClient})
	msc = metadata.NewDeviceServiceClient(
		types.EndpointParams{
			ServiceKey:  clients.CoreMetaDataServiceKey,
			Path:        clients.ApiDeviceServiceRoute,
			UseRegistry: registryClient != nil,
			Url:         Configuration.Clients["Metadata"].Url() + clients.ApiDeviceRoute,
			Interval:    Configuration.Service.ClientMonitor,
		},
		endpoint.Endpoint{RegistryClient: &registryClient})

	// Create the messaging client
	msgClient, err := messaging.NewMessageClient(
		msgTypes.MessageBusConfig{
			PublishHost: msgTypes.HostInfo{
				Host:     Configuration.MessageQueue.Host,
				Port:     Configuration.MessageQueue.Port,
				Protocol: Configuration.MessageQueue.Protocol,
			},
			Type: Configuration.MessageQueue.Type,
		})

	if err != nil {
		loggingClient.Error(fmt.Sprintf("failed to create messaging client: %s", err.Error()))
		return false
	}

	chEvents := make(chan interface{}, 100)
	// initialize event handlers
	initEventHandlers(loggingClient, chEvents, mdc)

	dic.Update(di.ServiceConstructorMap{
		dataContainer.MetadataDeviceClientName: func(get di.Get) interface{} {
			return mdc
		},
		dataContainer.MessagingClientName: func(get di.Get) interface{} {
			return msgClient
		},
		dataContainer.EventsChannelName: func(get di.Get) interface{} {
			return chEvents
		}})

	return true
}
