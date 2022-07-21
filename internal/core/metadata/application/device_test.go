//
// Copyright (C) 2022 Intel
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/edgexfoundry/go-mod-messaging/v2/messaging/mocks"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

func TestPublishDeviceSystemEvent(t *testing.T) {
	mockClient := &mocks.MessageClient{}
	lc := logger.NewMockClient()

	expectedDevice := models.Device{
		Name:        "Camera-Device",
		Id:          uuid.NewString(),
		ServiceName: "Device-onvif-camera",
		ProfileName: "onvif-camera",
	}

	tests := []struct {
		Name          string
		Action        string
		PubError      bool
		ClientMissing bool
	}{
		{"Device Add", common.DeviceSystemEventActionAdd, false, false},
		{"Device Update", common.DeviceSystemEventActionUpdate, false, false},
		{"Device Delete", common.DeviceSystemEventActionDelete, false, false},
		{"Publish Error", common.DeviceSystemEventActionAdd, true, false},
		{"Client Missing Error", common.DeviceSystemEventActionAdd, false, true},
	}

	pubErrMsg := errors.New("publish failed")

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			validatePublishCallFunc := func(envelope types.MessageEnvelope, topic string) error {
				assert.Equal(t, common.ContentTypeJSON, envelope.ContentType)
				assert.NotEmpty(t, envelope.CorrelationID)
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

			if test.PubError {
				mockClient.On("Publish", mock.Anything, mock.Anything).Return(pubErrMsg).Once()
			} else {
				mockClient.On("Publish", mock.Anything, mock.Anything).Return(validatePublishCallFunc).Once()
			}

			dic := di.NewContainer(di.ServiceConstructorMap{})

			if !test.ClientMissing {
				dic.Update(di.ServiceConstructorMap{
					container.MessagingClientName: func(get di.Get) interface{} {
						return mockClient
					},
				})
			}

			err := publishDeviceSystemEvent(test.Action, expectedDevice, lc, dic)

			if test.PubError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), pubErrMsg.Error())
				return
			} else if test.ClientMissing {
				require.Error(t, err)
				assert.Equal(t, noMessagingClientError, err)
				return
			}

			expectedTopic := fmt.Sprintf("%s/%s/%s/%s/%s/%s",
				common.SystemEventsPublishTopicPrefix,
				common.CoreMetaDataServiceKey,
				common.DeviceSystemEventType,
				test.Action,
				expectedDevice.ServiceName,
				expectedDevice.ProfileName)

			if !test.ClientMissing {
				mockClient.AssertCalled(t, "Publish", mock.Anything, expectedTopic)
			}
		})
	}
}
