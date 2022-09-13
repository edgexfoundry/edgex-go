//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package external

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces/mocks"
	lcMocks "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-messaging/v2/messaging/mocks"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/core/command/config"
	"github.com/edgexfoundry/edgex-go/internal/core/command/container"
	mqttMocks "github.com/edgexfoundry/edgex-go/internal/core/command/controller/messaging/mocks"
)

const (
	mockHost = "127.0.0.1"
	mockPort = 66666

	testProfileName       = "testProfile"
	testResourceName      = "testResource"
	testDeviceName        = "testDevice"
	testDeviceServiceName = "testService"

	testQueryRequestTopic        = "unittest/#"
	testQueryAllExample          = "unittest/all"
	testQueryByDeviceNameExample = "unittest/testDevice"
	testQueryResponseTopic       = "unittest/response"

	testCommandRequestTopic        = "unittest/external/request/#"
	testCommandRequestExample      = "unittest/external/request/testDevice/testCommand/get"
	testCommandResponseTopicPrefix = "unittest/external/response/"
)

func TestOnConnectHandler(t *testing.T) {
	lc := &lcMocks.LoggingClient{}
	lc.On("Errorf", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				MessageQueue: config.MessageQueue{
					Required: false,
					External: bootstrapConfig.ExternalMQTTInfo{
						Topics: map[string]string{
							RequestQueryTopic:          testQueryRequestTopic,
							ResponseQueryTopic:         testQueryResponseTopic,
							RequestCommandTopic:        testCommandRequestTopic,
							ResponseCommandTopicPrefix: testCommandResponseTopicPrefix,
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
			token := &mqttMocks.Token{}
			token.On("Wait").Return(true)
			if tt.expectedSucceed {
				token.On("Error").Return(nil)
			} else {
				token.On("Error").Return(errors.New("error"))
			}

			client := &mqttMocks.Client{}
			client.On("Subscribe", testQueryRequestTopic, byte(0), mock.Anything).Return(token)
			client.On("Subscribe", testCommandRequestTopic, byte(0), mock.Anything).Return(token)

			fn := OnConnectHandler(dic)
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
	lc.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return(nil)
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
		name              string
		requestQueryTopic string
		payload           types.MessageEnvelope
		expectedError     bool
	}{
		{"valid - query all", testQueryAllExample, validPayload, false},
		{"valid - query by device name", testQueryByDeviceNameExample, validPayload, false},
		{"invalid - invalid request json payload", testQueryByDeviceNameExample, invalidRequestPayload, true},
		{"invalid - invalid query parameters", testQueryAllExample, invalidQueryParamsPayload, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			message := &mqttMocks.Message{}
			message.On("Payload").Return(payloadBytes)
			message.On("Topic").Return(tt.requestQueryTopic)

			token := &mqttMocks.Token{}
			token.On("Wait").Return(true)
			token.On("Error").Return(nil)

			client := &mqttMocks.Client{}
			client.On("Publish", testQueryResponseTopic, byte(0), true, mock.Anything).Return(token)

			fn := commandQueryHandler(testQueryResponseTopic, 0, true, dic)
			fn(client, message)
			lc.AssertCalled(t, "Debugf", mock.Anything, mock.Anything, mock.Anything)
			if tt.expectedError {
				lc.AssertCalled(t, "Error", mock.Anything)
			}
		})
	}
}

func Test_commandRequestHandler(t *testing.T) {
	deviceResponse := responses.DeviceResponse{
		BaseResponse: commonDTO.NewBaseResponse("", "", http.StatusOK),
		Device: dtos.Device{
			Name:        testDeviceName,
			ProfileName: testProfileName,
			ServiceName: testDeviceServiceName,
		},
	}
	deviceServiceResponse := responses.DeviceServiceResponse{
		BaseResponse: commonDTO.NewBaseResponse("", "", http.StatusOK),
		Service: dtos.DeviceService{
			Name: testDeviceServiceName,
		},
	}

	lc := &lcMocks.LoggingClient{}
	lc.On("Error", mock.Anything).Return(nil)
	lc.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	dc := &clientMocks.DeviceClient{}
	dc.On("DeviceByName", context.Background(), testDeviceName).Return(deviceResponse, nil)
	dsc := &clientMocks.DeviceServiceClient{}
	dsc.On("DeviceServiceByName", context.Background(), testDeviceServiceName).Return(deviceServiceResponse, nil)
	client := &mocks.MessageClient{}
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
					Required: true,
					Internal: bootstrapConfig.MessageBusInfo{
						Topics: map[string]string{
							RequestTopicPrefix: "unittest/internal/request/",
						},
					},
					External: bootstrapConfig.ExternalMQTTInfo{
						QoS:    0,
						Retain: true,
						Topics: map[string]string{
							ResponseCommandTopicPrefix: testCommandResponseTopicPrefix,
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

	tests := []struct {
		name                string
		commandRequestTopic string
		payload             types.MessageEnvelope
		expectedError       bool
	}{
		{"valid", testCommandRequestExample, validPayload, false},
		{"invalid - invalid request json payload", testCommandRequestExample, invalidRequestPayload, true},
		{"invalid - unknown command method", "unittest/request/testDevice/testCommand/invalid", validPayload, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			message := &mqttMocks.Message{}
			message.On("Payload").Return(payloadBytes)
			message.On("Topic").Return(tt.commandRequestTopic)

			token := &mqttMocks.Token{}
			token.On("Wait").Return(true)
			token.On("Error").Return(nil)

			mqttClient := &mqttMocks.Client{}
			mqttClient.On("Publish", mock.Anything, byte(0), true, mock.Anything).Return(token)

			fn := commandRequestHandler(dic)
			fn(mqttClient, message)
			if tt.expectedError {
				lc.AssertCalled(t, "Error", mock.Anything)
			} else {
				client.AssertCalled(t, "Publish", tt.payload, mock.Anything)
			}
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
