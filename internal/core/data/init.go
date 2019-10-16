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

	"github.com/edgexfoundry/edgex-go/internal/core/data/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"

	"github.com/edgexfoundry/go-mod-messaging/messaging"
	msgTypes "github.com/edgexfoundry/go-mod-messaging/pkg/types"
)

// Global variables
var Configuration = &ConfigurationStruct{}
var dbClient interfaces.DBClient
var LoggingClient logger.LoggingClient

// TODO: Refactor names in separate PR: See comments on PR #1133
var chEvents chan interface{} // A channel for "domain events" sourced from event operations
var msgClient messaging.MessageClient
var mdc metadata.DeviceClient
var msc metadata.DeviceServiceClient

var httpErrorHandler errorconcept.ErrorHandler

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the data service.
func BootstrapHandler(wg *sync.WaitGroup, ctx context.Context, startupTimer startup.Timer, dic *di.Container) bool {
	// update global variables.
	LoggingClient = container.LoggingClientFrom(dic.Get)
	dbClient = container.DBClientFrom(dic.Get)

	httpErrorHandler = errorconcept.NewErrorHandler(LoggingClient)

	// initialize clients required by service.
	registryClient := container.RegistryFrom(dic.Get)
	mdc = metadata.NewDeviceClient(
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
	var err error
	msgClient, err = messaging.NewMessageClient(
		msgTypes.MessageBusConfig{
			PublishHost: msgTypes.HostInfo{
				Host:     Configuration.MessageQueue.Host,
				Port:     Configuration.MessageQueue.Port,
				Protocol: Configuration.MessageQueue.Protocol,
			},
			Type: Configuration.MessageQueue.Type,
		})

	if err != nil {
		LoggingClient.Error(fmt.Sprintf("failed to create messaging client: %s", err.Error()))
	}

	// initialize event handlers
	chEvents = make(chan interface{}, 100)
	initEventHandlers()

	return true
}
