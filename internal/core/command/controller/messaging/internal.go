//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/edgexfoundry/edgex-go/internal/core/command/container"
)

func SubscribeCommandResponses(ctx context.Context, router MessagingRouter, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	messageBusInfo := container.ConfigurationFrom(dic.Get).MessageQueue
	internalResponseTopic := messageBusInfo.Internal.Topics[ResponseTopic]

	messages := make(chan types.MessageEnvelope)
	messageErrors := make(chan error)
	topics := []types.TopicChannel{
		{
			Topic:    internalResponseTopic,
			Messages: messages,
		},
	}

	messageBus := bootstrapContainer.MessagingClientFrom(dic.Get)
	err := messageBus.Subscribe(topics, messageErrors)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	qos := messageBusInfo.External.QoS
	retain := messageBusInfo.External.Retain
	externalMQTT := bootstrapContainer.ExternalMQTTMessagingClientFrom(dic.Get)
	go func() {
		for {
			select {
			case <-ctx.Done():
				lc.Infof("Exiting waiting for MessageBus '%s' topic messages", internalResponseTopic)
				return
			case err = <-messageErrors:
				lc.Error(err.Error())
			case msgEnvelope := <-messages:
				lc.Debugf("Command response received on message queue. Topic: %s, Correlation-id: %s ", msgEnvelope.ReceivedTopic, msgEnvelope.CorrelationID)

				responseTopic, external, err := router.ResponseTopic(msgEnvelope.RequestID)
				if err != nil {
					lc.Errorf("Received RequestEnvelope with unknown RequestId %s", msgEnvelope.RequestID)
					continue
				}

				if external {
					publishMessage(externalMQTT, responseTopic, qos, retain, msgEnvelope, lc)
				}
			}
		}
	}()

	return nil
}
