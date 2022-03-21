//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"

	"github.com/edgexfoundry/go-mod-messaging/v2/messaging"
	msgTypes "github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
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
				Service: bootstrapConfig.ServiceInfo{
					MaxResultCount: 20,
				},
			}
		},
		container.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		container.MessagingClientName: func(get di.Get) interface{} {
			return msgClient
		},
	})
}
