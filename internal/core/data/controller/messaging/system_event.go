//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/core/data/application"
	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

func SubscribeSystemEvents(ctx context.Context, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := dataContainer.ConfigurationFrom(dic.Get)
	messageBusInfo := dataContainer.ConfigurationFrom(dic.Get).MessageBus

	// device deletion event edgex/system-events/core-metadata/device/delete/<device name>/<device profile name>
	deviceDeletionSystemEventTopic := common.NewPathBuilder().EnableNameFieldEscape(configuration.Service.EnableNameFieldEscape).
		SetPath(messageBusInfo.GetBaseTopicPrefix()).SetPath(common.SystemEventPublishTopic).SetPath(common.CoreMetaDataServiceKey).
		SetPath(common.DeviceSystemEventType).SetPath("#").BuildPath()
	lc.Infof("Subscribing to System Events on topic: %s", deviceDeletionSystemEventTopic)

	messages := make(chan types.MessageEnvelope, 1)
	messageErrors := make(chan error, 1)
	topics := []types.TopicChannel{
		{
			Topic:    deviceDeletionSystemEventTopic,
			Messages: messages,
		},
	}

	messageBus := bootstrapContainer.MessagingClientFrom(dic.Get)
	err := messageBus.Subscribe(topics, messageErrors)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				lc.Infof("Exiting waiting for MessageBus '%s' topic messages", deviceDeletionSystemEventTopic)
				return
			case err = <-messageErrors:
				lc.Error(err.Error())
			case msgEnvelope := <-messages:
				lc.Debugf("System event received on message queue. Topic: %s, Correlation-id: %s", msgEnvelope.ReceivedTopic, msgEnvelope.CorrelationID)
				var systemEvent dtos.SystemEvent
				systemEvent, err = types.GetMsgPayload[dtos.SystemEvent](msgEnvelope)
				if err != nil {
					lc.Errorf("failed to JSON decoding system event: %s", err.Error())
					continue
				}

				switch systemEvent.Type {
				case common.DeviceSystemEventType:
					err = deviceSystemEventAction(systemEvent, dic)
					if err != nil {
						lc.Error(err.Error(), common.CorrelationHeader, msgEnvelope.CorrelationID)
					}
				}
			}
		}
	}()

	return nil
}

func deviceSystemEventAction(systemEvent dtos.SystemEvent, dic *di.Container) error {
	var device dtos.Device
	err := systemEvent.DecodeDetails(&device)
	if err != nil {
		return fmt.Errorf("failed to decode %s system event details: %s", systemEvent.Type, err.Error())
	}

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	deviceStore := dataContainer.DeviceStoreFrom(dic.Get)
	switch systemEvent.Action {
	case common.SystemEventActionAdd:
		deviceStore.Add(dtos.ToDeviceModel(device))

	case common.SystemEventActionUpdate:
		deviceStore.Remove(device.Name)
		deviceStore.Add(dtos.ToDeviceModel(device))

	case common.SystemEventActionDelete:
		deviceStore.Remove(device.Name)

		if !dataContainer.ConfigurationFrom(dic.Get).Writable.EventPurge {
			return nil
		}
		lc.Debugf("Device '%s' is deleted, try to remove related events and readings.", device.Name)
		err = application.CoreDataAppFrom(dic.Get).DeleteEventsByDeviceName(device.Name, dic)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
		lc.Debugf("Events and readings are removed for the Device '%s'.", device.Name)
	}
	return nil
}
