//
// Copyright (C) 2022 Intel
// Copyright (C) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"encoding/json"
	goErrors "errors"
	"fmt"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/config"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	mocks2 "github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-messaging/v3/messaging/mocks"
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
)

func TestPublishSystemEvent(t *testing.T) {
	TestDeviceProfileName := "onvif-camera"
	TestDeviceServiceName := "Device-onvif-camera"
	TestDeviceName := "Camera-Device"

	expectedDevice := dtos.Device{
		Name:        TestDeviceName,
		Id:          uuid.NewString(),
		ServiceName: TestDeviceServiceName,
		ProfileName: TestDeviceProfileName,
	}

	expectedDeviceProfile := dtos.DeviceProfile{
		DeviceProfileBasicInfo: dtos.DeviceProfileBasicInfo{
			Id:   uuid.NewString(),
			Name: TestDeviceProfileName,
		},
	}

	expectedCorrelationID := uuid.NewString()
	expectedPublishTopicPrefix := "edgex/system-events"

	tests := []struct {
		Name          string
		Type          string
		Action        string
		Owner         string
		PubError      bool
		ClientMissing bool
	}{
		{"Device Add", common.DeviceSystemEventType, common.SystemEventActionAdd, TestDeviceServiceName, false, false},
		{"Device Update", common.DeviceSystemEventType, common.SystemEventActionUpdate, TestDeviceServiceName, false, false},
		{"Device Delete", common.DeviceSystemEventType, common.SystemEventActionDelete, TestDeviceServiceName, false, false},
		{"Device Profile Add", common.DeviceProfileSystemEventType, common.SystemEventActionAdd, common.CoreMetaDataServiceKey, false, false},
		{"Device Profile Update", common.DeviceProfileSystemEventType, common.SystemEventActionUpdate, TestDeviceServiceName, false, false},
		{"Device Profile Delete", common.DeviceProfileSystemEventType, common.SystemEventActionDelete, common.CoreMetaDataServiceKey, false, false},
		{"Client Missing Error", common.DeviceSystemEventType, common.SystemEventActionAdd, TestDeviceServiceName, false, true},
		{"Publish Error", common.DeviceSystemEventType, common.SystemEventActionAdd, TestDeviceServiceName, true, false},
	}

	pubErrMsg := errors.NewCommonEdgeXWrapper(goErrors.New("publish failed"))
	mockLogger := &mocks2.LoggingClient{}

	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{}
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return mockLogger
		},
	})

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			validatePublishCallFunc := func(envelope types.MessageEnvelope, topic string) error {
				assert.Equal(t, common.ContentTypeJSON, envelope.ContentType)
				assert.Equal(t, expectedCorrelationID, envelope.CorrelationID)
				require.NotEmpty(t, envelope.Payload)
				systemEvent := dtos.SystemEvent{}
				err := json.Unmarshal(envelope.Payload, &systemEvent)
				require.NoError(t, err)

				switch test.Type {
				case common.DeviceSystemEventType:
					actualDevice := dtos.Device{}
					err = systemEvent.DecodeDetails(&actualDevice)
					require.NoError(t, err)
					assert.Equal(t, expectedDevice.Name, actualDevice.Name)
					assert.Equal(t, expectedDevice.Id, actualDevice.Id)
					assert.Equal(t, expectedDevice.ServiceName, actualDevice.ServiceName)
					assert.Equal(t, expectedDevice.ProfileName, actualDevice.ProfileName)
				case common.DeviceProfileSystemEventType:
					actualDeviceProfile := dtos.DeviceProfile{}
					err = systemEvent.DecodeDetails(&actualDeviceProfile)
					require.NoError(t, err)
					assert.Equal(t, expectedDeviceProfile.Name, actualDeviceProfile.Name)
					assert.Equal(t, expectedDeviceProfile.Id, actualDeviceProfile.Id)
				}

				assert.Equal(t, common.ApiVersion, systemEvent.ApiVersion)
				assert.Equal(t, test.Type, systemEvent.Type)
				assert.Equal(t, test.Action, systemEvent.Action)
				assert.Equal(t, common.CoreMetaDataServiceKey, systemEvent.Source)
				assert.Equal(t, test.Owner, systemEvent.Owner)
				assert.NotZero(t, systemEvent.Timestamp)

				return nil
			}

			mockClient := &mocks.MessageClient{}
			mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

			if test.PubError {
				mockClient.On("Publish", mock.Anything, mock.Anything).Return(pubErrMsg)
				mockLogger.On("Errorf", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
			} else {
				mockClient.On("Publish", mock.Anything, mock.Anything).Return(validatePublishCallFunc)
			}

			if test.ClientMissing {
				mockLogger.On("Errorf", mock.Anything, mock.Anything, mock.Anything).Return()
				dic.Update(di.ServiceConstructorMap{
					bootstrapContainer.MessagingClientName: func(get di.Get) interface{} {
						return nil
					},
				})
			} else {
				dic.Update(di.ServiceConstructorMap{
					bootstrapContainer.MessagingClientName: func(get di.Get) interface{} {
						return mockClient
					},
				})
			}

			// Use CBOR to make sure publisher overrides with JSON properly
			// lint:ignore SA1029 legacy
			// nolint:staticcheck // See golangci-lint #741
			ctx := context.WithValue(context.Background(), common.ContentType, common.ContentTypeCBOR)
			// lint:ignore SA1029 legacy
			// nolint:staticcheck // See golangci-lint #741
			ctx = context.WithValue(ctx, common.CorrelationHeader, expectedCorrelationID)
			var expectedDetails any
			switch test.Type {
			case common.DeviceSystemEventType:
				expectedDetails = expectedDevice
			case common.DeviceProfileSystemEventType:
				expectedDetails = expectedDeviceProfile
			}

			publishSystemEvent(test.Type, test.Action, test.Owner, expectedDetails, ctx, dic)

			if test.ClientMissing {
				mockLogger.AssertCalled(t, "Errorf", mock.Anything, mock.Anything, noMessagingClientError)
				return
			}

			expectedTopic := fmt.Sprintf("%s/%s/%s/%s/%s/%s",
				expectedPublishTopicPrefix,
				common.CoreMetaDataServiceKey,
				test.Type,
				test.Action,
				test.Owner,
				expectedDevice.ProfileName)
			mockClient.AssertCalled(t, "Publish", mock.Anything, expectedTopic)

			if test.PubError {
				mockLogger.AssertCalled(t, "Errorf", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, pubErrMsg)
			}
		})
	}
}
