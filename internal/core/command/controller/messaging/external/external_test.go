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
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/core/command/config"
	"github.com/edgexfoundry/edgex-go/internal/core/command/container"
	"github.com/edgexfoundry/edgex-go/internal/core/command/controller/messaging/mocks"
)

const (
	mockHost = "127.0.0.1"
	mockPort = 66666

	testProfileName  = "testProfile"
	testResourceName = "testResource"
	testDeviceName   = "testDevice"

	testRequestQueryTopic             = "unittest/#"
	testRequestQueryAllTopic          = "unittest/all"
	testRequestQueryByDeviceNameTopic = "unittest/testDevice"
	testResponseTopic                 = "unittest/response"
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
							RequestQueryTopic:  testRequestQueryTopic,
							ResponseQueryTopic: testResponseTopic,
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
			client.On("Subscribe", testRequestQueryTopic, byte(0), mock.Anything).Return(token)

			fn := OnConnectHandler(dic)
			fn(client)

			client.AssertCalled(t, "Subscribe", testRequestQueryTopic, byte(0), mock.Anything)
			if !tt.expectedSucceed {
				lc.AssertCalled(t, "Errorf", mock.Anything, mock.Anything, mock.Anything)
			}
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

	validPayload := testPayload()
	invalidRequestPayload := testPayload()
	invalidRequestPayload.ApiVersion = "v1"
	invalidQueryParamsPayload := testPayload()
	invalidQueryParamsPayload.QueryParams[common.Offset] = "invalid"

	tests := []struct {
		name              string
		requestQueryTopic string
		payload           types.MessageEnvelope
		expectedError     bool
	}{
		{"valid - query all", testRequestQueryAllTopic, validPayload, false},
		{"valid - query by device name", testRequestQueryByDeviceNameTopic, validPayload, false},
		{"invalid - invalid request json payload", testRequestQueryByDeviceNameTopic, invalidRequestPayload, true},
		{"invalid - invalid query parameters", testRequestQueryAllTopic, invalidQueryParamsPayload, true},
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

			client := &mocks.Client{}
			client.On("Publish", testResponseTopic, byte(0), true, mock.Anything).Return(token)

			fn := commandQueryHandler(testResponseTopic, 0, true, dic)
			fn(client, message)
			lc.AssertCalled(t, "Debugf", mock.Anything, mock.Anything, mock.Anything)
			if tt.expectedError {
				lc.AssertCalled(t, "Error", mock.Anything)
			}
		})
	}
}

func testPayload() types.MessageEnvelope {
	payload := types.NewMessageEnvelopeForRequest(nil, nil)

	return payload
}
