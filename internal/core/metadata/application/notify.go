//
// Copyright (C) 2021-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
)

// validateDeviceCallback invoke device service's validation function for validating new or updated device
func validateDeviceCallback(device dtos.Device, dic *di.Container) errors.EdgeX {
	configuration := container.ConfigurationFrom(dic.Get)
	messagingClient := bootstrapContainer.MessagingClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	requestTimeout, err := time.ParseDuration(configuration.Service.RequestTimeout)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to parse service.RequestTimeout", err)
	}

	// reusing AddDeviceRequest here as it contains the protocols field and opens up
	// to other validation beyond protocols if ever needed
	addDeviceRequest := requests.NewAddDeviceRequest(device)
	requestBytes, err := json.Marshal(addDeviceRequest)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to JSON encoding AddDeviceRequest", err)
	}

	baseTopic := configuration.MessageBus.GetBaseTopicPrefix()
	requestTopic := common.BuildTopic(baseTopic, device.ServiceName, common.ValidateDeviceSubscribeTopic)
	responseTopicPrefix := common.BuildTopic(baseTopic, common.ResponseTopic, device.ServiceName)
	requestEnvelope := types.NewMessageEnvelopeForRequest(requestBytes, nil)

	lc.Debugf("Sending Device Validation request for device=%s, CorrelationId=%s to topic: %s", device.Name, requestEnvelope.CorrelationID, requestTopic)
	lc.Debugf("Waiting for Device Validation response on topic: %s/%s", responseTopicPrefix, requestEnvelope.RequestID)

	res, err := messagingClient.Request(requestEnvelope, requestTopic, responseTopicPrefix, requestTimeout)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServiceUnavailable, fmt.Sprintf("Error sending request to topic '%s'", requestTopic), err)
	} else if res.ErrorCode == 1 {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("Device %s validation failed: %s", device.Name, res.Payload), nil)
	}

	lc.Debugf("Received Device Validation response for device=%s, CorrelationId=%s on topic: %s", device.Name, res.CorrelationID, res.ReceivedTopic)

	return nil
}

func publishUpdateDeviceProfileSystemEvent(profileDTO dtos.DeviceProfile, ctx context.Context, dic *di.Container) {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	devices, _, err := DevicesByProfileName(0, -1, profileDTO.Name, dic)
	if err != nil {
		lc.Errorf("fail to query associated devices by deviceProfile name %s, err: %v", profileDTO.Name, err)
		return
	}

	//Publish general system event regardless of associated devices
	publishSystemEvent(common.DeviceProfileSystemEventType, common.SystemEventActionUpdate, common.CoreMetaDataServiceKey, profileDTO, ctx, dic)
	// Publish system event for each device service
	dsMap := make(map[string]bool)
	for _, d := range devices {
		if _, ok := dsMap[d.ServiceName]; ok {
			// skip the invoked device service
			continue
		}
		dsMap[d.ServiceName] = true

		publishSystemEvent(common.DeviceProfileSystemEventType, common.SystemEventActionUpdate, d.ServiceName, profileDTO, ctx, dic)
	}
}

func publishSystemEvent(eventType, action, owner string, dto any, ctx context.Context, dic *di.Container) {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	systemEvent := dtos.NewSystemEvent(eventType, action, common.CoreMetaDataServiceKey, owner, nil, dto)
	messagingClient := bootstrapContainer.MessagingClientFrom(dic.Get)
	if messagingClient == nil {
		lc.Errorf("unable to publish '%s' System Event: %v", eventType, noMessagingClientError)
		return
	}

	var profileName, detailName string
	switch eventType {
	case common.DeviceSystemEventType:
		if device, ok := dto.(dtos.Device); ok {
			profileName = device.ProfileName
			detailName = device.Name
		} else {
			lc.Errorf("can not convert to device DTO")
			return
		}
	case common.DeviceProfileSystemEventType:
		if profile, ok := dto.(dtos.DeviceProfile); ok {
			profileName = profile.Name
			detailName = profile.Name
		} else {
			lc.Errorf("can not convert to device profile DTO")
			return
		}
	case common.ProvisionWatcherSystemEventType:
		if pw, ok := dto.(dtos.ProvisionWatcher); ok {
			profileName = pw.DiscoveredDevice.ProfileName
			detailName = pw.Name
		} else {
			lc.Errorf("can not convert to provision watcher DTO")
		}
	case common.DeviceServiceSystemEventType:
		if service, ok := dto.(dtos.DeviceService); ok {
			detailName = service.Name
		} else {
			lc.Errorf("can not convert to device service DTO")
			return
		}
	default:
		lc.Errorf("unrecognized system event details")
		return
	}

	publishTopic := common.BuildTopic(
		container.ConfigurationFrom(dic.Get).MessageBus.GetBaseTopicPrefix(),
		common.SystemEventPublishTopic,
		systemEvent.Source,
		systemEvent.Type,
		systemEvent.Action,
		systemEvent.Owner,
	)

	if profileName != "" {
		publishTopic = common.BuildTopic(
			publishTopic,
			profileName,
		)
	}

	payload, _ := json.Marshal(systemEvent)
	envelope := types.NewMessageEnvelope(payload, ctx)
	// Correlation ID and Content type are set by the above factory function from the context of the request that
	// triggered this System Event. We'll keep that Correlation ID, but need to make sure the Content Type is set appropriate
	// for how the payload was encoded above.
	envelope.ContentType = common.ContentTypeJSON

	if err := messagingClient.Publish(envelope, publishTopic); err != nil {
		lc.Errorf("unable to publish '%s' System Event for %s '%s' to topic '%s': %v", action, eventType, detailName, publishTopic, err)
		return
	}

	lc.Debugf("Published the '%s' System Event for %s '%s' to topic '%s'", action, eventType, detailName, publishTopic)
}
