//
// Copyright (C) 2022-2026 IOTech Ltd
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

			fn := OnConnectHandler(time.Second*10, dic)
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
	client.On("Request", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)
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

			sem := make(chan struct{}, defaultMaxConcurrentExternalCommands)
			fn := commandRequestHandler(time.Second*10, sem, dic)
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

			expectedInternalRequestTopic := common.BuildTopic(baseTopic, common.CoreCommandDeviceRequestPublishTopic, testDeviceServiceName, testDeviceName, testCommandName, testMethod)
			expectedInternalResponseTopicPrefix := common.BuildTopic(baseTopic, common.ResponseTopic, testDeviceServiceName)
			// Request now runs in a goroutine — wait for it.
			require.Eventually(t, func() bool {
				return client.AssertCalled(new(testing.T), "Request", tt.payload, expectedInternalRequestTopic, expectedInternalResponseTopicPrefix, mock.Anything)
			}, time.Second, 10*time.Millisecond)
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

// newCommandRequestDIC builds a DIC and MessageClient that runs runRequest before responding to
// Request. Shared helper for the concurrency tests below.
func newCommandRequestDIC(t *testing.T, runRequest func(args mock.Arguments)) (*di.Container, *internalMessagingMocks.MessageClient) {
	t.Helper()

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
		Service:      dtos.DeviceService{Name: testDeviceServiceName},
	}

	lc := &lcMocks.LoggingClient{}
	lc.On("Error", mock.Anything).Return(nil)
	lc.On("Errorf", mock.Anything, mock.Anything).Return(nil)
	lc.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	lc.On("Warn", mock.Anything).Return(nil)
	lc.On("Warnf", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	dc := &clientMocks.DeviceClient{}
	dc.On("DeviceByName", context.Background(), testDeviceName).Return(deviceResponse, nil)
	dsc := &clientMocks.DeviceServiceClient{}
	dsc.On("DeviceServiceByName", context.Background(), testDeviceServiceName).Return(deviceServiceResponse, nil)

	msgClient := &internalMessagingMocks.MessageClient{}
	call := msgClient.On("Request", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	if runRequest != nil {
		call = call.Run(runRequest)
	}
	call.Return(&types.MessageEnvelope{}, nil)

	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Service:    bootstrapConfig.ServiceInfo{Host: mockHost, Port: mockPort, MaxResultCount: 20},
				MessageBus: bootstrapConfig.MessageBusInfo{BaseTopicPrefix: "edgex"},
				ExternalMQTT: bootstrapConfig.ExternalMQTTInfo{
					QoS: 0, Retain: true,
					Topics: map[string]string{
						common.CommandResponseTopicPrefixKey: testExternalCommandResponseTopicPrefix,
					},
				},
			}
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} { return lc },
		bootstrapContainer.DeviceClientName:           func(get di.Get) interface{} { return dc },
		bootstrapContainer.DeviceServiceClientName:    func(get di.Get) interface{} { return dsc },
		bootstrapContainer.MessagingClientName:        func(get di.Get) interface{} { return msgClient },
	})

	return dic, msgClient
}

// Test_commandRequestHandler_secondRunsWhileFirstBlocked proves the fix: a second external command
// can complete the internal Request while the first is still blocked, i.e. no head-of-line blocking
// on the Paho callback goroutine.
func Test_commandRequestHandler_secondRunsWhileFirstBlocked(t *testing.T) {
	unblockFirst := make(chan struct{})
	var phase atomic.Int32

	dic, msgClient := newCommandRequestDIC(t, func(_ mock.Arguments) {
		if phase.Add(1) == 1 {
			<-unblockFirst
		}
	})

	payloadBytes, err := json.Marshal(testCommandRequestPayload())
	require.NoError(t, err)

	token := &mocks.Token{}
	token.On("Wait").Return(true)
	token.On("Error").Return(nil)
	mqttClient := &mocks.Client{}
	mqttClient.On("Publish", mock.Anything, byte(0), true, mock.Anything).Return(token)

	sem := make(chan struct{}, 4)
	handler := commandRequestHandler(10*time.Second, sem, dic)

	msg := func() *mocks.Message {
		m := &mocks.Message{}
		m.On("Payload").Return(payloadBytes)
		m.On("Topic").Return(testExternalCommandRequestTopicExample)
		return m
	}

	handler(mqttClient, msg())
	require.Eventually(t, func() bool { return phase.Load() >= 1 }, 2*time.Second, 5*time.Millisecond,
		"first Request should have started")

	done := make(chan struct{})
	go func() {
		handler(mqttClient, msg())
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		close(unblockFirst)
		t.Fatal("second MQTT handler invocation blocked while first Request was in flight — HOL regression")
	}

	require.Eventually(t, func() bool { return phase.Load() >= 2 }, 2*time.Second, 5*time.Millisecond,
		"second Request should run while first is blocked")
	close(unblockFirst)

	// Drain so workers can release the semaphore before the test ends.
	require.Eventually(t, func() bool { return len(sem) == 0 }, 2*time.Second, 5*time.Millisecond)
	_ = msgClient
}

// Test_commandRequestHandler_busyOnOverload proves the backpressure path: when the in-flight
// budget is exhausted, further requests are rejected inline with an error envelope and the
// internal Request is NOT invoked again.
func Test_commandRequestHandler_busyOnOverload(t *testing.T) {
	blockAll := make(chan struct{})
	var entered atomic.Int32

	dic, msgClient := newCommandRequestDIC(t, func(_ mock.Arguments) {
		entered.Add(1)
		<-blockAll
	})

	payloadBytes, err := json.Marshal(testCommandRequestPayload())
	require.NoError(t, err)

	token := &mocks.Token{}
	token.On("Wait").Return(true)
	token.On("Error").Return(nil)
	var publishCount atomic.Int32
	mqttClient := &mocks.Client{}
	mqttClient.On("Publish", mock.Anything, byte(0), true, mock.Anything).
		Run(func(_ mock.Arguments) { publishCount.Add(1) }).
		Return(token)

	sem := make(chan struct{}, 1) // capacity-1 so the second request must overload
	handler := commandRequestHandler(10*time.Second, sem, dic)

	msg := func() *mocks.Message {
		m := &mocks.Message{}
		m.On("Payload").Return(payloadBytes)
		m.On("Topic").Return(testExternalCommandRequestTopicExample)
		return m
	}

	handler(mqttClient, msg())
	require.Eventually(t, func() bool { return entered.Load() == 1 }, 2*time.Second, 5*time.Millisecond,
		"first internal Request should have started")

	handler(mqttClient, msg()) // second — should be rejected inline

	// Internal Request must still have been called only once; the second was load-shed.
	msgClient.AssertNumberOfCalls(t, "Request", 1)
	// The rejection publishes a busy envelope on the response topic.
	require.Eventually(t, func() bool { return publishCount.Load() >= 1 }, 2*time.Second, 5*time.Millisecond,
		"busy response should be published for the rejected request")

	close(blockAll)
	require.Eventually(t, func() bool { return len(sem) == 0 }, 2*time.Second, 5*time.Millisecond)
}
