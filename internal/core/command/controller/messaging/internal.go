//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"strings"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/edgexfoundry/edgex-go/internal/core/command/container"
)

func SubscribeCommandResponses(ctx context.Context, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	messageBusInfo := container.ConfigurationFrom(dic.Get).MessageQueue
	internalResponseTopic := messageBusInfo.Internal.Topics[ResponseTopic]
	externalResponseTopicPrefix := messageBusInfo.External.Topics[ResponseCommandTopicPrefix]

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
			case err := <-messageErrors:
				lc.Error(err.Error())
			case msgEnvelope := <-messages:
				lc.Debugf("Command response received on message queue. Topic: %s, Correlation-id: %s ", internalResponseTopic, msgEnvelope.CorrelationID)

				// expected internal command response topic scheme: #/<service-name>/<device-name>/<command-name>/<method>
				topicLevels := strings.Split(msgEnvelope.ReceivedTopic, "/")
				length := len(topicLevels)
				if length < 4 {
					lc.Error("Failed to parse and construct command response topic scheme, expected request topic scheme: '#/<service-name>/<device-name>/<command-name>/<method>'")
					continue
				}

				// expected external command response topic scheme: #/<device-name>/<command-name>/<method>
				deviceName := topicLevels[length-3]
				commandName := topicLevels[length-2]
				method := topicLevels[length-1]
				externalResponseTopic := strings.Join([]string{externalResponseTopicPrefix, deviceName, commandName, method}, "/")
				publishMessage(externalMQTT, externalResponseTopic, qos, retain, msgEnvelope, lc)
			}
		}
	}()

	return nil
}
