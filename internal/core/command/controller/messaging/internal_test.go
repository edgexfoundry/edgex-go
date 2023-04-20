// Copyright (C) 2023 Intel Inc.
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	config2 "github.com/edgexfoundry/go-mod-bootstrap/v3/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	mocks2 "github.com/edgexfoundry/go-mod-core-contracts/v3/clients/interfaces/mocks"
	lcMocks "github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	"github.com/edgexfoundry/go-mod-messaging/v3/messaging/mocks"
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/core/command/config"
	"github.com/edgexfoundry/edgex-go/internal/core/command/container"
)

var expectedResponseTopicPrefix = "edgex/response"
var expectedProfileName = "TestProfile"
var baseTopic = "edgex"

func TestSubscribeCommandRequests(t *testing.T) {
	wg := sync.WaitGroup{}
	expectedServiceName := "device-simple"
	expectedRequestId := uuid.NewString()
	expectedCorrelationId := uuid.NewString()
	expectedDevice := "device1"
	expectedResource := "resource"
	expectedMethod := "get"
	expectedDeviceResponseTopicPrefix := strings.Join([]string{expectedResponseTopicPrefix, expectedServiceName}, "/")
	expectedCommandResponseTopic := strings.Join([]string{expectedResponseTopicPrefix, common.CoreCommandServiceKey, expectedRequestId}, "/")
	expectedCommandRequestSubscribeTopic := common.BuildTopic(baseTopic, common.CoreCommandRequestSubscribeTopic)
	expectedCommandRequestReceivedTopic := common.BuildTopic(strings.Replace(expectedCommandRequestSubscribeTopic, "/#", "", 1),
		expectedServiceName, expectedDevice, expectedResource, expectedMethod)
	expectedDeviceCommandRequestRequestTopic := common.BuildTopic(baseTopic, common.CoreCommandDeviceRequestPublishTopic, expectedServiceName, expectedDevice, expectedResource, expectedMethod)
	mockLogger := &lcMocks.LoggingClient{}
	mockDeviceClient := &mocks2.DeviceClient{}
	mockDeviceProfileClient := &mocks2.DeviceProfileClient{}
	mockDeviceServiceClient := &mocks2.DeviceServiceClient{}
	mockMessaging := &mocks.MessageClient{}

	mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	mockMessaging.On("Subscribe", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		topics := args.Get(0).([]types.TopicChannel)
		require.Len(t, topics, 1)
		require.Equal(t, expectedCommandRequestSubscribeTopic, topics[0].Topic)
		wg.Add(1)
		go func() {
			defer wg.Done()
			topics[0].Messages <- types.MessageEnvelope{
				RequestID:     expectedRequestId,
				CorrelationID: expectedCorrelationId,
				ReceivedTopic: expectedCommandRequestReceivedTopic,
			}
			time.Sleep(time.Second * 1)
		}()
	}).Return(nil)

	mockMessaging.On("Request", mock.Anything, expectedDeviceCommandRequestRequestTopic, expectedDeviceResponseTopicPrefix, mock.Anything).Run(func(args mock.Arguments) {
	}).Return(&types.MessageEnvelope{
		RequestID:     expectedRequestId,
		CorrelationID: expectedCorrelationId,
		ContentType:   common.ContentTypeJSON,
		Payload:       []byte("This is my payload"),
	}, nil)

	mockMessaging.On("Publish", mock.Anything, expectedCommandResponseTopic).Run(func(args mock.Arguments) {
		response := args.Get(0).(types.MessageEnvelope)
		assert.Equal(t, expectedRequestId, response.RequestID)
		assert.Equal(t, expectedCorrelationId, response.CorrelationID)
		assert.Equal(t, common.ContentTypeJSON, response.ContentType)
		assert.NotZero(t, len(response.Payload))
	}).Return(nil)

	mockDeviceClient.On("DeviceByName", mock.Anything, expectedDevice).Return(
		responses.DeviceResponse{
			Device: dtos.Device{
				ProfileName: expectedProfileName,
				ServiceName: expectedServiceName,
			},
		},
		nil)

	mockDeviceServiceClient.On("DeviceServiceByName", mock.Anything, expectedServiceName).Return(
		responses.DeviceServiceResponse{
			Service: dtos.DeviceService{
				Name: expectedServiceName,
			},
		},
		nil)

	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				MessageBus: config2.MessageBusInfo{
					BaseTopicPrefix: "edgex",
				},
			}
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return mockLogger
		},
		bootstrapContainer.MessagingClientName: func(get di.Get) interface{} {
			return mockMessaging
		},
		bootstrapContainer.DeviceClientName: func(get di.Get) interface{} {
			return mockDeviceClient
		},
		bootstrapContainer.DeviceProfileClientName: func(get di.Get) interface{} {
			return mockDeviceProfileClient
		},
		bootstrapContainer.DeviceServiceClientName: func(get di.Get) interface{} {
			return mockDeviceServiceClient
		},
	})

	err := SubscribeCommandRequests(context.Background(), time.Second*5, dic)
	require.NoError(t, err)

	wg.Wait()

	mockMessaging.AssertExpectations(t)
}

func TestSubscribeCommandQueryRequests(t *testing.T) {
	wg := sync.WaitGroup{}
	expectedRequestId := uuid.NewString()
	expectedCorrelationId := uuid.NewString()
	expectedResponseTopic := strings.Join([]string{expectedResponseTopicPrefix, common.CoreCommandServiceKey, expectedRequestId}, "/")

	tests := []struct {
		Name               string
		ExpectedDeviceName string
	}{
		{"By Device", "Device1"},
		{"All Devices", "All"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {

			mockLogger := &lcMocks.LoggingClient{}
			mockDeviceClient := &mocks2.DeviceClient{}
			mockDeviceProfileClient := &mocks2.DeviceProfileClient{}
			mockMessaging := &mocks.MessageClient{}

			mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			mockLogger.On("Infof", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			mockLogger.On("Errorf", mock.Anything).Run(func(args mock.Arguments) {
				require.Fail(t, "Errorf not expected")
			})
			mockLogger.On("Error", mock.Anything).Run(func(args mock.Arguments) {
				require.Fail(t, "Error not expected")
			})

			expectedSubscribeTopic := common.BuildTopic(baseTopic, common.CoreCommandQueryRequestSubscribeTopic)
			expectedReceivedTopic := common.BuildTopic(strings.Replace(expectedSubscribeTopic, "/#", "", 1), test.ExpectedDeviceName)

			mockDeviceClient.On("DeviceByName", mock.Anything, test.ExpectedDeviceName).Return(
				responses.DeviceResponse{
					Device: dtos.Device{
						ProfileName: expectedProfileName,
					},
				},
				nil)

			mockDeviceClient.On("AllDevices", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
				responses.MultiDevicesResponse{
					Devices: []dtos.Device{
						{
							ProfileName: expectedProfileName,
						},
					},
				},
				nil)

			mockDeviceProfileClient.On("DeviceProfileByName", mock.Anything, expectedProfileName).Return(
				responses.DeviceProfileResponse{},
				nil)

			mockMessaging.On("Subscribe", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				topics := args.Get(0).([]types.TopicChannel)
				require.Len(t, topics, 1)
				require.Equal(t, expectedSubscribeTopic, topics[0].Topic)
				wg.Add(1)
				go func() {
					defer wg.Done()
					topics[0].Messages <- types.MessageEnvelope{
						RequestID:     expectedRequestId,
						CorrelationID: expectedCorrelationId,
						ReceivedTopic: expectedReceivedTopic,
					}
					time.Sleep(time.Second * 1)
				}()
			}).Return(nil)
			mockMessaging.On("Publish", mock.Anything, expectedResponseTopic).Run(func(args mock.Arguments) {
				response := args.Get(0).(types.MessageEnvelope)
				assert.Equal(t, expectedRequestId, response.RequestID)
				assert.Equal(t, expectedCorrelationId, response.CorrelationID)
				assert.Equal(t, common.ContentTypeJSON, response.ContentType)
				assert.NotZero(t, len(response.Payload))
			}).Return(nil)

			dic := di.NewContainer(di.ServiceConstructorMap{
				container.ConfigurationName: func(get di.Get) interface{} {
					return &config.ConfigurationStruct{
						MessageBus: config2.MessageBusInfo{
							BaseTopicPrefix: "edgex",
						},
					}
				},
				bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
					return mockLogger
				},
				bootstrapContainer.MessagingClientName: func(get di.Get) interface{} {
					return mockMessaging
				},
				bootstrapContainer.DeviceClientName: func(get di.Get) interface{} {
					return mockDeviceClient
				},
				bootstrapContainer.DeviceProfileClientName: func(get di.Get) interface{} {
					return mockDeviceProfileClient
				},
			})

			err := SubscribeCommandQueryRequests(context.Background(), dic)
			require.NoError(t, err)

			wg.Wait()

			mockMessaging.AssertExpectations(t)

		})
	}
}
