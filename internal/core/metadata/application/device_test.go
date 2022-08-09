//
// Copyright (C) 2022 Intel
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
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	mocks2 "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-messaging/v2/messaging/mocks"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

func TestPublishDeviceSystemEvent(t *testing.T) {
	expectedDevice := models.Device{
		Name:        "Camera-Device",
		Id:          uuid.NewString(),
		ServiceName: "Device-onvif-camera",
		ProfileName: "onvif-camera",
	}

	expectedCorrelationID := uuid.NewString()
	expectedPublishTopicPrefix := "events/system-event"

	tests := []struct {
		Name          string
		Action        string
		PubError      bool
		ClientMissing bool
	}{
		{"Device Add", common.DeviceSystemEventActionAdd, false, false},
		{"Device Update", common.DeviceSystemEventActionUpdate, false, false},
		{"Device Delete", common.DeviceSystemEventActionDelete, false, false},
		{"Client Missing Error", common.DeviceSystemEventActionAdd, false, true},
		{"Publish Error", common.DeviceSystemEventActionAdd, true, false},
	}

	pubErrMsg := errors.NewCommonEdgeXWrapper(goErrors.New("publish failed"))

	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				MessageQueue: bootstrapConfig.MessageBusInfo{
					PublishTopicPrefix: expectedPublishTopicPrefix},
			}
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

				assert.Equal(t, common.ApiVersion, systemEvent.ApiVersion)
				assert.Equal(t, common.DeviceSystemEventType, systemEvent.Type)
				assert.Equal(t, test.Action, systemEvent.Action)
				assert.Equal(t, common.CoreMetaDataServiceKey, systemEvent.Source)
				assert.Equal(t, expectedDevice.ServiceName, systemEvent.Owner)
				assert.NotZero(t, systemEvent.Timestamp)

				actualDevice := dtos.Device{}
				err = systemEvent.DecodeDetails(&actualDevice)
				require.NoError(t, err)

				assert.Equal(t, expectedDevice.Name, actualDevice.Name)
				assert.Equal(t, expectedDevice.Id, actualDevice.Id)
				assert.Equal(t, expectedDevice.ServiceName, actualDevice.ServiceName)
				assert.Equal(t, expectedDevice.ProfileName, actualDevice.ProfileName)
				return nil
			}

			mockClient := &mocks.MessageClient{}
			mockLogger := &mocks2.LoggingClient{}
			mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

			if test.PubError {
				mockClient.On("Publish", mock.Anything, mock.Anything).Return(pubErrMsg)
				mockLogger.On("Errorf", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
			} else {
				mockClient.On("Publish", mock.Anything, mock.Anything).Return(validatePublishCallFunc)
			}

			if test.ClientMissing {
				// TODO: Change to Errorf in EdgeX 3.0
				mockLogger.On("Warnf", mock.Anything, mock.Anything).Return()
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

			publishDeviceSystemEvent(test.Action, expectedDevice.ServiceName, expectedDevice, ctx, mockLogger, dic)

			if test.ClientMissing {
				// TODO: Change to Errorf in EdgeX 3.0
				mockLogger.AssertCalled(t, "Warnf", mock.Anything, noMessagingClientError)
				return
			}

			expectedTopic := fmt.Sprintf("%s/%s/%s/%s/%s/%s",
				expectedPublishTopicPrefix,
				common.CoreMetaDataServiceKey,
				common.DeviceSystemEventType,
				test.Action,
				expectedDevice.ServiceName,
				expectedDevice.ProfileName)
			mockClient.AssertCalled(t, "Publish", mock.Anything, expectedTopic)

			if test.PubError {
				mockLogger.AssertCalled(t, "Errorf", mock.Anything, mock.Anything, mock.Anything, mock.Anything, pubErrMsg)
			}
		})
	}
}
