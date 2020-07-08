//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	v2DataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/v2/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/error"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/edgexfoundry/go-mod-messaging/messaging"
	msgTypes "github.com/edgexfoundry/go-mod-messaging/pkg/types"
)

// NewMockDIC function returns a mock bootstrap di Container
func NewMockDIC() *di.Container {
	msgClient, _ := messaging.NewMessageClient(msgTypes.MessageBusConfig{
		PublishHost: msgTypes.HostInfo{
			Host:     "*",
			Protocol: "tcp",
			Port:     5563,
		},
		Type: "zero",
	})

	return di.NewContainer(di.ServiceConstructorMap{
		dataContainer.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					PersistData: true,
				},
			}
		},
		v2DataContainer.ErrorHandlerName: func(get di.Get) interface{} {
			return error.NewErrorHandler(logger.NewMockClient())
		},
		container.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		v2DataContainer.MetadataDeviceClientName: func(get di.Get) interface{} {
			return NewMockDeviceClient()
		},
		dataContainer.MessagingClientName: func(get di.Get) interface{} {
			return msgClient
		},
	})
}
