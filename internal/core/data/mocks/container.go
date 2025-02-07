//
// Copyright (C) 2020-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/go-mod-messaging/v4/messaging/mocks"
	"github.com/stretchr/testify/mock"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

// NewMockDIC function returns a mock bootstrap di Container
func NewMockDIC() *di.Container {
	msgClient := &mocks.MessageClient{}
	msgClient.On("PublishWithSizeLimit", mock.Anything, mock.Anything, mock.Anything).Return(nil)

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
