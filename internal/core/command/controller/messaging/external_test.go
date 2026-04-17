//
// Copyright (C) 2022-2025 IOTech Ltd
// Copyright (C) 2023 Intel Inc.
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces/mocks"
	lcMocks "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	internalMessagingMocks "github.com/edgexfoundry/go-mod-messaging/v4/messaging/mocks"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"

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
)

func TestOnConnectHandler(t *testing.T) {
	lc := &lcMocks.LoggingClient{}
	lc.On("Errorf", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	lc.On("Debugf", mock.Anything, mock.Anything).Return(nil)
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				ExternalMQTT: bootstrapConfig.ExternalMQTTInfo{
					Topics: map[string]string{
						common.CommandRequestTopicKey:        testExternalCommandRequestTopic,
						common.CommandResponseTopicPrefixKey: testExternalCommandResponseTopicPrefix,
						common.CommandQueryRequestTopicKey:   testQueryRequestTopic,
						common.CommandQueryResponseTopicKey:  testQueryResponseTopic,
					},
					QoS:    0,
					Retain: true,
				},
				ExternalCommandQueue: config.ExternalCommandQueue{
					MaxConcurrentExternalCommands:  4,
					MaxQueuedExternalCommands:      8,
					OverloadPublishChannelCapacity: 4,
					ShutdownTimeout:                "30s",
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

			fn := OnConnectHandler(context.Background(), time.Second*10, dic)
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
				{
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
			{
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
				ExternalMQTT: bootstrapConfig.ExternalMQTTInfo{
					QoS:    0,
					Retain: true,
					Topics: map[string]string{
						common.CommandQueryResponseTopicKey: testQueryResponseTopic,
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
					lc.AssertCalled(t, "Errorf", mock.Anything, mock.Anything)
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

	expectedResponse := &types.MessageEnvelope{}
	baseTopic := "edgex"

	lc := &lcMocks.LoggingClient{}
	lc.On("Error", mock.Anything).Return(nil)
	lc.On("Errorf", mock.Anything, mock.Anything).Return(nil)
	lc.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	lc.On("Warn", mock.Anything).Return(nil)
	dc := &clientMocks.DeviceClient{}
	dc.On("DeviceByName", context.Background(), testDeviceName).Return(deviceResponse, nil)
	dc.On("DeviceByName", context.Background(), unknownDevice).Return(responses.DeviceResponse{}, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, "unknown device", nil))
	dc.On("DeviceByName", context.Background(), unknownServiceDevice).Return(unknownServiceDeviceResponse, nil)
	dsc := &clientMocks.DeviceServiceClient{}
	dsc.On("DeviceServiceByName", context.Background(), testDeviceServiceName).Return(deviceServiceResponse, nil)
	dsc.On("DeviceServiceByName", context.Background(), unknownService).Return(responses.DeviceServiceResponse{}, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, "unknown device service", nil))
	client := &internalMessagingMocks.MessageClient{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Service: bootstrapConfig.ServiceInfo{
					Host:           mockHost,
					Port:           mockPort,
					MaxResultCount: 20,
				},
				MessageBus: bootstrapConfig.MessageBusInfo{
					BaseTopicPrefix: "edgex",
				},
				ExternalMQTT: bootstrapConfig.ExternalMQTTInfo{
					QoS:    0,
					Retain: true,
					Topics: map[string]string{
						common.CommandResponseTopicPrefixKey: testExternalCommandResponseTopicPrefix,
					},
				},
				ExternalCommandQueue: config.ExternalCommandQueue{
					MaxConcurrentExternalCommands:  8,
					MaxQueuedExternalCommands:      16,
					OverloadPublishChannelCapacity: 4,
					ShutdownTimeout:                "30s",
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

			requestDone := make(chan struct{}, 1)
			if !tt.expectedError {
				client.On("Request", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) { requestDone <- struct{}{} }).
					Return(expectedResponse, nil).Once()
			}

			proc := newExternalCommandProcessor(context.Background(), time.Second*10, dic, externalCommandLimitsFromConfig(container.ConfigurationFrom(dic.Get).ExternalCommandQueue))
			proc.setMQTTClient(mqttClient)
			proc.ensureStarted()
			fn := proc.commandRequestMQTTHandler()
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

			select {
			case <-requestDone:
			case <-time.After(3 * time.Second):
				t.Fatal("timed out waiting for internal MessageBus Request")
			}

			expectedInternalRequestTopic := common.BuildTopic(baseTopic, common.CoreCommandDeviceRequestPublishTopic, testDeviceServiceName, testDeviceName, testCommandName, testMethod)
			expectedInternalResponseTopicPrefix := common.BuildTopic(baseTopic, common.ResponseTopic, testDeviceServiceName)
			client.AssertCalled(t, "Request", tt.payload, expectedInternalRequestTopic, expectedInternalResponseTopicPrefix, mock.Anything)
		})
	}
}

func testCommandQueryPayload() types.MessageEnvelope {
	payload := types.NewMessageEnvelopeForRequest(nil, nil)

	return payload
}

func testCommandRequestPayload() types.MessageEnvelope {
	payload := types.NewMessageEnvelopeForRequest(nil, map[string]string{
		"ds-pushevent":   common.ValueTrue,
		"ds-returnevent": common.ValueTrue,
	})

	return payload
}

// Test_externalCommand_secondRequestWhileFirstBlocked ensures a second external command can complete
// internal Request while the first Request is still blocked (no MQTT callback head-of-line blocking).
func Test_externalCommand_secondRequestWhileFirstBlocked(t *testing.T) {
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
	expectedResponse := &types.MessageEnvelope{}

	lc := &lcMocks.LoggingClient{}
	lc.On("Error", mock.Anything).Return(nil)
	lc.On("Errorf", mock.Anything, mock.Anything).Return(nil)
	lc.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	dc := &clientMocks.DeviceClient{}
	dc.On("DeviceByName", context.Background(), testDeviceName).Return(deviceResponse, nil)
	dsc := &clientMocks.DeviceServiceClient{}
	dsc.On("DeviceServiceByName", context.Background(), testDeviceServiceName).Return(deviceServiceResponse, nil)

	unblockFirst := make(chan struct{})
	var requestPhase atomic.Int32

	msgClient := &internalMessagingMocks.MessageClient{}
	msgClient.On("Request", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			n := requestPhase.Add(1)
			if n == 1 {
				<-unblockFirst
			}
		}).Return(expectedResponse, nil)

	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Service: bootstrapConfig.ServiceInfo{
					Host: mockHost, Port: mockPort, MaxResultCount: 20,
				},
				MessageBus: bootstrapConfig.MessageBusInfo{BaseTopicPrefix: "edgex"},
				ExternalMQTT: bootstrapConfig.ExternalMQTTInfo{
					QoS: 0, Retain: true,
					Topics: map[string]string{
						common.CommandResponseTopicPrefixKey: testExternalCommandResponseTopicPrefix,
					},
				},
				ExternalCommandQueue: config.ExternalCommandQueue{
					MaxConcurrentExternalCommands:  4,
					MaxQueuedExternalCommands:      8,
					OverloadPublishChannelCapacity: 4,
					ShutdownTimeout:                "30s",
				},
			}
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} { return lc },
		bootstrapContainer.DeviceClientName:           func(get di.Get) interface{} { return dc },
		bootstrapContainer.DeviceServiceClientName:    func(get di.Get) interface{} { return dsc },
		bootstrapContainer.MessagingClientName:        func(get di.Get) interface{} { return msgClient },
	})

	payload := testCommandRequestPayload()
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	token := &mocks.Token{}
	token.On("Wait").Return(true)
	token.On("Error").Return(nil)
	mqttClient := &mocks.Client{}
	mqttClient.On("Publish", mock.Anything, byte(0), true, mock.Anything).Return(token)

	proc := newExternalCommandProcessor(context.Background(), time.Second*10, dic,
		externalCommandLimitsFromConfig(container.ConfigurationFrom(dic.Get).ExternalCommandQueue))
	proc.setMQTTClient(mqttClient)
	proc.ensureStarted()
	handler := proc.commandRequestMQTTHandler()

	message1 := &mocks.Message{}
	message1.On("Payload").Return(append([]byte(nil), payloadBytes...))
	message1.On("Topic").Return(testExternalCommandRequestTopicExample)
	handler(mqttClient, message1)

	require.Eventually(t, func() bool { return requestPhase.Load() >= 1 }, 2*time.Second, 5*time.Millisecond,
		"first Request should have started")

	message2 := &mocks.Message{}
	message2.On("Payload").Return(append([]byte(nil), payloadBytes...))
	message2.On("Topic").Return(testExternalCommandRequestTopicExample)

	done2 := make(chan struct{})
	go func() {
		handler(mqttClient, message2)
		close(done2)
	}()

	select {
	case <-done2:
	case <-time.After(3 * time.Second):
		close(unblockFirst)
		t.Fatal("second MQTT handler invocation did not return")
	}

	require.Eventually(t, func() bool { return requestPhase.Load() >= 2 }, 2*time.Second, 5*time.Millisecond,
		"second Request should run while first is blocked")
	close(unblockFirst)
	time.Sleep(100 * time.Millisecond)
}

// Test_externalCommand_queueFullOverload exercises non-blocking enqueue when the jobs channel is full.
func Test_externalCommand_queueFullOverload(t *testing.T) {
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
	expectedResponse := &types.MessageEnvelope{}

	lc := &lcMocks.LoggingClient{}
	lc.On("Error", mock.Anything).Return(nil)
	lc.On("Errorf", mock.Anything, mock.Anything).Return(nil)
	lc.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	lc.On("Warn", mock.Anything).Return(nil)
	dc := &clientMocks.DeviceClient{}
	dc.On("DeviceByName", context.Background(), testDeviceName).Return(deviceResponse, nil)
	dsc := &clientMocks.DeviceServiceClient{}
	dsc.On("DeviceServiceByName", context.Background(), testDeviceServiceName).Return(deviceServiceResponse, nil)

	blockAll := make(chan struct{})
	var requestEntered atomic.Int32
	msgClient := &internalMessagingMocks.MessageClient{}
	msgClient.On("Request", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			requestEntered.Add(1)
			<-blockAll
		}).Return(expectedResponse, nil)

	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Service: bootstrapConfig.ServiceInfo{
					Host: mockHost, Port: mockPort, MaxResultCount: 20,
				},
				MessageBus: bootstrapConfig.MessageBusInfo{BaseTopicPrefix: "edgex"},
				ExternalMQTT: bootstrapConfig.ExternalMQTTInfo{
					QoS: 0, Retain: true,
					Topics: map[string]string{
						common.CommandResponseTopicPrefixKey: testExternalCommandResponseTopicPrefix,
					},
				},
				ExternalCommandQueue: config.ExternalCommandQueue{
					MaxConcurrentExternalCommands:  1,
					MaxQueuedExternalCommands:      1,
					OverloadPublishChannelCapacity: 4,
					ShutdownTimeout:                "30s",
				},
			}
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} { return lc },
		bootstrapContainer.DeviceClientName:           func(get di.Get) interface{} { return dc },
		bootstrapContainer.DeviceServiceClientName:    func(get di.Get) interface{} { return dsc },
		bootstrapContainer.MessagingClientName:        func(get di.Get) interface{} { return msgClient },
	})

	payload := testCommandRequestPayload()
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	token := &mocks.Token{}
	token.On("Wait").Return(true)
	token.On("Error").Return(nil)
	var publishCount atomic.Int32
	mqttClient := &mocks.Client{}
	mqttClient.On("Publish", mock.Anything, byte(0), true, mock.Anything).
		Run(func(args mock.Arguments) { publishCount.Add(1) }).
		Return(token)

	proc := newExternalCommandProcessor(context.Background(), time.Second*10, dic,
		externalCommandLimitsFromConfig(container.ConfigurationFrom(dic.Get).ExternalCommandQueue))
	proc.setMQTTClient(mqttClient)
	proc.ensureStarted()
	handler := proc.commandRequestMQTTHandler()

	send := func() {
		message := &mocks.Message{}
		message.On("Payload").Return(append([]byte(nil), payloadBytes...))
		message.On("Topic").Return(testExternalCommandRequestTopicExample)
		handler(mqttClient, message)
	}

	send()
	require.Eventually(t, func() bool { return requestEntered.Load() >= 1 },
		2*time.Second, 5*time.Millisecond, "first internal Request should start")
	send()
	time.Sleep(20 * time.Millisecond)
	send()

	msgClient.AssertNumberOfCalls(t, "Request", 1)
	// Third message was load-shed; overload response is published via the dedicated publisher.
	require.Eventually(t, func() bool { return publishCount.Load() >= 1 },
		2*time.Second, 10*time.Millisecond, "overload response publish")

	close(blockAll)
	time.Sleep(50 * time.Millisecond)
}
