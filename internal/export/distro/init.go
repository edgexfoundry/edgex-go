//
// Copyright (c) 2018 Tencent
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"context"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"

	"github.com/edgexfoundry/go-mod-messaging/messaging"
	msgTypes "github.com/edgexfoundry/go-mod-messaging/pkg/types"

	"github.com/edgexfoundry/go-mod-registry/registry"
)

// Global variables
var LoggingClient logger.LoggingClient
var ec coredata.EventClient
var Configuration = &ConfigurationStruct{}
var messageErrors chan error
var messageEnvelopes chan msgTypes.MessageEnvelope

// initializeClients creates the clients required by the service.
func initializeClients(useRegistry bool, registryClient registry.Client) (messaging.MessageClient, error) {
	ec = coredata.NewEventClient(
		types.EndpointParams{
			ServiceKey:  clients.CoreDataServiceKey,
			Path:        clients.ApiEventRoute,
			UseRegistry: useRegistry,
			Url:         Configuration.Clients["CoreData"].Url() + clients.ApiEventRoute,
			Interval:    Configuration.Service.ClientMonitor,
		},
		endpoint.Endpoint{RegistryClient: &registryClient})

	// Create the messaging client
	return messaging.NewMessageClient(
		msgTypes.MessageBusConfig{
			SubscribeHost: msgTypes.HostInfo{
				Host:     Configuration.MessageQueue.Host,
				Port:     Configuration.MessageQueue.Port,
				Protocol: Configuration.MessageQueue.Protocol,
			},
			Type: Configuration.MessageQueue.Type,
		})
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the export-distro service.
func BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	// update global variables.
	LoggingClient = container.LoggingClientFrom(dic.Get)

	// initialize clients required by service.
	registryClient := container.RegistryFrom(dic.Get)
	messageClient, err := initializeClients(registryClient != nil, registryClient)
	if err != nil {
		LoggingClient.Error("failed to create messaging client: " + err.Error())
		return false
	}

	// initialize Messaging
	messageErrors, messageEnvelopes, err = initMessaging(messageClient)
	if err != nil {
		LoggingClient.Error(err.Error())
		return false
	}

	Loop(wg, ctx)

	return true
}
