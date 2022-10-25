//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces/mocks"
	lcMocks "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	internalMessagingMocks "github.com/edgexfoundry/go-mod-messaging/v2/messaging/mocks"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/edgexfoundry/edgex-go/internal/core/command/config"
	"github.com/edgexfoundry/edgex-go/internal/core/command/container"
	"github.com/edgexfoundry/edgex-go/internal/core/command/controller/messaging/mocks"
)

const (
	mockHost = "127.0.0.1"
	mockPort = 66666

	testProfileName       = "testProfile"
	testResourceName      = "testResource"
	testDeviceName        = "testDevice"
	testDeviceServiceName = "testService"
	testCommandName       = "testCommand"
	testMethod            = "get"

	testQueryRequestTopic        = "unittest/#"
	testQueryAllExample          = "unittest/all"
	testQueryByDeviceNameExample = "unittest/testDevice"
	testQueryResponseTopic       = "unittest/response"

	testExternalCommandRequestTopic        = "unittest/external/request/#"
	testExternalCommandRequestTopicExample = "unittest/external/request/testDevice/testCommand/get"
	testExternalCommandResponseTopicPrefix = "unittest/external/response"

	testInternalCommandRequestTopicPrefix = "unittest/internal/request"
)

func TestOnConnectHandler(t *testing.T) {
	mockRouter := &mocks.MessagingRouter{}
	lc := &lcMocks.LoggingClient{}
	lc.On("Errorf", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	lc.On("Debugf", mock.Anything, mock.Anything).Return(nil)
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				MessageQueue: config.MessageQueue{
					External: bootstrapConfig.ExternalMQTTInfo{
						Topics: map[string]string{
							QueryRequestTopic:          testQueryRequestTopic,
							QueryResponseTopic:         testQueryResponseTopic,
							CommandRequestTopic:        testExternalCommandRequestTopic,
							CommandResponseTopicPrefix: testExternalCommandResponseTopicPrefix,
						},
						QoS:    0,
						Retain: true,
					},
				},
			}
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
	})

	tests := []struct {
		name            string
		expectedSucceed bool
	}{
		{"subscribe succeed", true},
		{"subscribe fail", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &mocks.Token{}
			token.On("Wait").Return(true)
			if tt.expectedSucceed {
				token.On("Error").Return(nil)
			} else {
				token.On("Error").Return(errors.New("error"))
			}

			client := &mocks.Client{}
			client.On("Subscribe", testQueryRequestTopic, byte(0), mock.Anything).Return(token)
			client.On("Subscribe", testExternalCommandRequestTopic, byte(0), mock.Anything).Return(token)

			fn := OnConnectHandler(mockRouter, dic)
			fn(client)

			if tt.expectedSucceed {
				client.AssertNumberOfCalls(t, "Subscribe", 2)
				return
			}

			lc.AssertCalled(t, "Errorf", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func Test_commandQueryHandler(t *testing.T) {
	profileResponse := responses.DeviceProfileResponse{
		BaseResponse: commonDTO.NewBaseResponse("", "", http.StatusOK),
		Profile: dtos.DeviceProfile{
			DeviceProfileBasicInfo: dtos.DeviceProfileBasicInfo{
				Name: testProfileName,
			},
			DeviceResources: []dtos.DeviceResource{
				dtos.DeviceResource{
					Name: testResourceName,
					Properties: dtos.ResourceProperties{
						ValueType: common.ValueTypeString,
						ReadWrite: common.ReadWrite_RW,
					},
				},
			},
		},
	}
	deviceResponse := responses.DeviceResponse{
		BaseResponse: commonDTO.NewBaseResponse("", "", http.StatusOK),
		Device: dtos.Device{
			Name:        testDeviceName,
			ProfileName: testProfileName,
		},
	}
	allDevicesResponse := responses.MultiDevicesResponse{
		BaseWithTotalCountResponse: commonDTO.NewBaseWithTotalCountResponse("", "", http.StatusOK, 1),
		Devices: []dtos.Device{
			dtos.Device{
				Name:        testDeviceName,
				ProfileName: testProfileName,
			},
		},
	}

	lc := &lcMocks.LoggingClient{}
	lc.On("Error", mock.Anything).Return(nil)
	lc.On("Errorf", mock.Anything, mock.Anything).Return(nil)
	lc.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	lc.On("Warn", mock.Anything).Return(nil)
	dc := &clientMocks.DeviceClient{}
	dc.On("AllDevices", context.Background(), []string(nil), common.DefaultOffset, common.DefaultLimit).Return(allDevicesResponse, nil)
	dc.On("DeviceByName", context.Background(), testDeviceName).Return(deviceResponse, nil)
	dpc := &clientMocks.DeviceProfileClient{}
	dpc.On("DeviceProfileByName", context.Background(), testProfileName).Return(profileResponse, nil)
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Service: bootstrapConfig.ServiceInfo{
					Host:           mockHost,
					Port:           mockPort,
					MaxResultCount: 20,
				},
				MessageQueue: config.MessageQueue{
					Internal: bootstrapConfig.MessageBusInfo{},
					External: bootstrapConfig.ExternalMQTTInfo{
						QoS:    0,
						Retain: true,
						Topics: map[string]string{
							QueryResponseTopic: testQueryResponseTopic,
						},
					},
				},
			}
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
		bootstrapContainer.DeviceClientName: func(get di.Get) interface{} {
			return dc
		},
		bootstrapContainer.DeviceProfileClientName: func(get di.Get) interface{} {
			return dpc
		},
	})

	validPayload := testCommandQueryPayload()
	invalidRequestPayload := testCommandQueryPayload()
	invalidRequestPayload.ApiVersion = "v1"
	invalidQueryParamsPayload := testCommandQueryPayload()
	invalidQueryParamsPayload.QueryParams[common.Offset] = "invalid"

	tests := []struct {
		name                 string
		requestQueryTopic    string
		payload              types.MessageEnvelope
		expectedError        bool
		expectedPublishError bool
	}{
		{"valid - query all", testQueryAllExample, validPayload, false, false},
		{"valid - query by device name", testQueryByDeviceNameExample, validPayload, false, false},
		{"invalid - invalid request json payload", testQueryByDeviceNameExample, invalidRequestPayload, true, false},
		{"invalid - invalid query parameters", testQueryAllExample, invalidQueryParamsPayload, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			message := &mocks.Message{}
			message.On("Payload").Return(payloadBytes)
			message.On("Topic").Return(tt.requestQueryTopic)

			token := &mocks.Token{}
			token.On("Wait").Return(true)
			token.On("Error").Return(nil)

			mqttClient := &mocks.Client{}
			mqttClient.On("Publish", testQueryResponseTopic, byte(0), true, mock.Anything).Return(token)

			fn := commandQueryHandler(dic)
			fn(mqttClient, message)
			if tt.expectedError {
				if tt.expectedPublishError {
					lc.AssertCalled(t, "Error", mock.Anything)
					mqttClient.AssertCalled(t, "Publish", testQueryResponseTopic, byte(0), true, mock.Anything)
					return
				}
				lc.AssertCalled(t, "Warn", mock.Anything)
				return
			}

			mqttClient.AssertCalled(t, "Publish", testQueryResponseTopic, byte(0), true, mock.Anything)
			lc.AssertCalled(t, "Debugf", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func Test_commandRequestHandler(t *testing.T) {
	unknownDevice := "unknown-device"
	unknownServiceDevice := "unknownService-device"
	unknownService := "unknown-service"

	deviceResponse := responses.DeviceResponse{
		BaseResponse: commonDTO.NewBaseResponse("", "", http.StatusOK),
		Device: dtos.Device{
			Name:        testDeviceName,
			ProfileName: testProfileName,
			ServiceName: testDeviceServiceName,
		},
	}
	unknownServiceDeviceResponse := responses.DeviceResponse{
		BaseResponse: commonDTO.NewBaseResponse("", "", http.StatusOK),
		Device: dtos.Device{
			Name:        unknownServiceDevice,
			ProfileName: testProfileName,
			ServiceName: unknownService,
		},
	}
	deviceServiceResponse := responses.DeviceServiceResponse{
		BaseResponse: commonDTO.NewBaseResponse("", "", http.StatusOK),
		Service: dtos.DeviceService{
			Name: testDeviceServiceName,
		},
	}

	mockRouter := &mocks.MessagingRouter{}
	mockRouter.On("SetResponseTopic", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	lc := &lcMocks.LoggingClient{}
	lc.On("Error", mock.Anything).Return(nil)
	lc.On("Errorf", mock.Anything, mock.Anything).Return(nil)
	lc.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	lc.On("Warn", mock.Anything).Return(nil)
	dc := &clientMocks.DeviceClient{}
	dc.On("DeviceByName", context.Background(), testDeviceName).Return(deviceResponse, nil)
	dc.On("DeviceByName", context.Background(), unknownDevice).Return(responses.DeviceResponse{}, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, "unknown device", nil))
	dc.On("DeviceByName", context.Background(), unknownServiceDevice).Return(unknownServiceDeviceResponse, nil)
	dsc := &clientMocks.DeviceServiceClient{}
	dsc.On("DeviceServiceByName", context.Background(), testDeviceServiceName).Return(deviceServiceResponse, nil)
	dsc.On("DeviceServiceByName", context.Background(), unknownService).Return(responses.DeviceServiceResponse{}, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, "unknown device service", nil))
	client := &internalMessagingMocks.MessageClient{}
	client.On("Publish", mock.Anything, mock.Anything).Return(nil)
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Service: bootstrapConfig.ServiceInfo{
					Host:           mockHost,
					Port:           mockPort,
					MaxResultCount: 20,
				},
				MessageQueue: config.MessageQueue{
					Internal: bootstrapConfig.MessageBusInfo{
						Topics: map[string]string{
							DeviceRequestTopicPrefix: testInternalCommandRequestTopicPrefix,
						},
					},
					External: bootstrapConfig.ExternalMQTTInfo{
						QoS:    0,
						Retain: true,
						Topics: map[string]string{
							CommandResponseTopicPrefix: testExternalCommandResponseTopicPrefix,
						},
					},
				},
			}
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
		bootstrapContainer.DeviceClientName: func(get di.Get) interface{} {
			return dc
		},
		bootstrapContainer.DeviceServiceClientName: func(get di.Get) interface{} {
			return dsc
		},
		bootstrapContainer.MessagingClientName: func(get di.Get) interface{} {
			return client
		},
	})

	validPayload := testCommandRequestPayload()
	invalidRequestPayload := testCommandRequestPayload()
	invalidRequestPayload.ApiVersion = "v1"
	invalidQueryParamsPayload := testCommandQueryPayload()
	invalidQueryParamsPayload.QueryParams[common.PushEvent] = "invalid"
	invalidQueryParamsPayload.QueryParams[common.ReturnEvent] = "invalid"

	tests := []struct {
		name                 string
		commandRequestTopic  string
		payload              types.MessageEnvelope
		expectedError        bool
		expectedPublishError bool
	}{
		{"valid", testExternalCommandRequestTopicExample, validPayload, false, false},
		{"invalid - invalid request json payload", testExternalCommandRequestTopicExample, invalidRequestPayload, true, false},
		{"invalid - invalid request topic scheme", "unittest/invalid", validPayload, true, false},
		{"invalid - unrecognized command method", "unittest/request/testDevice/testCommand/invalid", validPayload, true, false},
		{"invalid - device not found", "unittest/request/unknown-device/testCommand/get", validPayload, true, true},
		{"invalid - device service not found", "unittest/request/unknownService-device/testCommand/get", validPayload, true, true},
		{"invalid - invalid device service reserved query parameters", testExternalCommandRequestTopicExample, invalidQueryParamsPayload, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			message := &mocks.Message{}
			message.On("Payload").Return(payloadBytes)
			message.On("Topic").Return(tt.commandRequestTopic)

			token := &mocks.Token{}
			token.On("Wait").Return(true)
			token.On("Error").Return(nil)

			mqttClient := &mocks.Client{}
			mqttClient.On("Publish", mock.Anything, byte(0), true, mock.Anything).Return(token)

			fn := commandRequestHandler(mockRouter, dic)
			fn(mqttClient, message)
			if tt.expectedError {
				if tt.expectedPublishError {
					lc.AssertCalled(t, "Error", mock.Anything)
					mqttClient.AssertCalled(t, "Publish", mock.Anything, byte(0), true, mock.Anything)
					return
				}
				lc.AssertCalled(t, "Warn", mock.Anything)
				return
			}

			expectedInternalRequestTopic := strings.Join([]string{testInternalCommandRequestTopicPrefix, testDeviceServiceName, testDeviceName, testCommandName, testMethod}, "/")
			client.AssertCalled(t, "Publish", tt.payload, expectedInternalRequestTopic)
		})
	}
}

func testCommandQueryPayload() types.MessageEnvelope {
	payload := types.NewMessageEnvelopeForRequest(nil, nil)

	return payload
}

func testCommandRequestPayload() types.MessageEnvelope {
	payload := types.NewMessageEnvelopeForRequest(nil, map[string]string{
		"ds-pushevent":   "yes",
		"ds-returnevent": "yes",
	})

	return payload
}
