//
// Copyright (C) 2021-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal/core/data/application"
	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// SubscribeEvents subscribes to events from message bus
func SubscribeEvents(ctx context.Context, dic *di.Container) errors.EdgeX {
	messageBusInfo := dataContainer.ConfigurationFrom(dic.Get).MessageBus
	lc := container.LoggingClientFrom(dic.Get)

	messageBus := container.MessagingClientFrom(dic.Get)

	messages := make(chan types.MessageEnvelope)
	messageErrors := make(chan error)

	app := application.CoreDataAppFrom(dic.Get)

	subscribeTopic := common.BuildTopic(messageBusInfo.GetBaseTopicPrefix(), common.CoreDataEventSubscribeTopic)

	topics := []types.TopicChannel{
		{
			Topic:    subscribeTopic,
			Messages: messages,
		},
	}

	err := messageBus.Subscribe(topics, messageErrors)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	// Log the topics to which core-data is subscribing
	for _, t := range topics {
		lc.Infof("Subscribed to topics: %s", t.Topic)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				lc.Infof("Exiting waiting for MessageBus '%s' topic messages", subscribeTopic)
				return
			case e := <-messageErrors:
				lc.Error(e.Error())
			case msgEnvelope := <-messages:
				lc.Debugf("Event received from MessageBus. Topic: %s, Correlation-id: %s", msgEnvelope.ReceivedTopic, msgEnvelope.CorrelationID)
				var event requests.AddEventRequest
				event, err = types.GetMsgPayload[requests.AddEventRequest](msgEnvelope)
				if err != nil {
					lc.Errorf("fail to unmarshal event, %v", err)
					break
				}
				err = validateEvent(msgEnvelope.ReceivedTopic, event.Event)
				if err != nil {
					lc.Error(err.Error())
					break
				}
				err = app.AddEvent(requests.AddEventReqToEventModel(event), ctx, dic)
				if err != nil {
					lc.Errorf("fail to persist the event, %v", err)
				}
			}
		}
	}()

	return nil
}

func validateEvent(messageTopic string, e dtos.Event) errors.EdgeX {
	// Parse messageTopic by the pattern `edgex/events/device/<device-service-name>/<device-profile-name>/<device-name>/<source-name>`
	fields := strings.Split(messageTopic, "/")

	// assumes a non-empty base topic with events/device/<device-service-name>/<device-profile-name>/<device-name>/<source-name>
	if len(fields) < 6 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("invalid message topic %s", messageTopic), nil)
	}

	len := len(fields)
	profileName, err := url.PathUnescape(fields[len-3])
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	deviceName, err := url.PathUnescape(fields[len-2])
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	sourceName, err := url.PathUnescape(fields[len-1])
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	// Check whether the event fields match the message topic
	if e.ProfileName != profileName {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("event's profileName %s mismatches with the name %s received in topic", e.ProfileName, profileName), nil)
	}
	if e.DeviceName != deviceName {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("event's deviceName %s mismatches with the name %s received in topic", e.DeviceName, deviceName), nil)
	}
	if e.SourceName != sourceName {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("event's sourceName %s mismatches with the name %s received in topic", e.SourceName, sourceName), nil)
	}
	return nil
}
