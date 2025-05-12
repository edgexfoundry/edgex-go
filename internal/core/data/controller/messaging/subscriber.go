//
// Copyright (C) 2021-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"

	"github.com/edgexfoundry/edgex-go/internal/core/data/application"
	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
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
				err = utils.ValidateEvent(msgEnvelope.ReceivedTopic, event.Event)
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
