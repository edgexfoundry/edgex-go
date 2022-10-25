//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/edgexfoundry/edgex-go/internal/core/command/container"
)

// SubscribeCommandResponses subscribes command responses from device services via internal MessageBus
func SubscribeCommandResponses(ctx context.Context, router MessagingRouter, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	messageBusInfo := container.ConfigurationFrom(dic.Get).MessageQueue
	internalResponseTopic := messageBusInfo.Internal.Topics[DeviceResponseTopic]

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
				lc.Debugf("Command response received on internal MessageBus. Topic: %s, Correlation-id: %s", msgEnvelope.ReceivedTopic, msgEnvelope.CorrelationID)

				responseTopic, external, err := router.ResponseTopic(msgEnvelope.RequestID)
				if err != nil {
					lc.Errorf("Received RequestEnvelope with unknown RequestId %s", msgEnvelope.RequestID)
					continue
				}

				// original request is from external MQTT
				if external {
					publishMessage(externalMQTT, responseTopic, qos, retain, msgEnvelope, lc)
					continue
				}

				// original request is from internal MessageBus
				err = messageBus.Publish(msgEnvelope, responseTopic)
				if err != nil {
					lc.Errorf("Could not publish to internal MessageBus topic '%s': %s", responseTopic, err.Error())
					continue
				}
				lc.Debugf("Command response sent to internal MessageBus. Topic: %s, Correlation-id: %s", responseTopic, msgEnvelope.CorrelationID)
			}
		}
	}()

	return nil
}

// SubscribeCommandRequests subscribes command requests from EdgeX service (e.g., Application Service)
// via internal MessageBus
func SubscribeCommandRequests(ctx context.Context, router MessagingRouter, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	messageBusInfo := container.ConfigurationFrom(dic.Get).MessageQueue
	internalRequestCommandTopic := messageBusInfo.Internal.Topics[CommandRequestTopic]

	messages := make(chan types.MessageEnvelope)
	messageErrors := make(chan error)
	topics := []types.TopicChannel{
		{
			Topic:    internalRequestCommandTopic,
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
				lc.Infof("Exiting waiting for MessageBus '%s' topic messages", internalRequestCommandTopic)
				return
			case err = <-messageErrors:
				lc.Error(err.Error())
			case requestEnvelope := <-messages:
				lc.Debugf("Command request received on internal MessageBus. Topic: %s, Correlation-id: %s", requestEnvelope.ReceivedTopic, requestEnvelope.CorrelationID)

				topicLevels := strings.Split(requestEnvelope.ReceivedTopic, "/")
				length := len(topicLevels)
				if length < 3 {
					lc.Error("Failed to parse and construct internal command response topic scheme, expected request topic scheme: '#/<device-name>/<command-name>/<method>'")
					lc.Warn("Not publishing error message back due to insufficient information on response topic")
					continue
				}

				// expected internal command request/response topic scheme: #/<device>/<command-name>/<method>
				deviceName := topicLevels[length-3]
				commandName := topicLevels[length-2]
				method := topicLevels[length-1]
				if !strings.EqualFold(method, "get") && !strings.EqualFold(method, "set") {
					lc.Errorf("Unknown request method: %s, only 'get' or 'set' is allowed", method)
					lc.Warn("Not publishing error message back due to insufficient information on response topic")
					continue
				}
				internalResponseTopic := strings.Join([]string{messageBusInfo.Internal.Topics[CommandResponseTopicPrefix], deviceName, commandName, method}, "/")

				deviceRequestTopic, err := validateRequestTopic(messageBusInfo.Internal.Topics[DeviceRequestTopicPrefix], deviceName, commandName, method, dic)
				if err != nil {
					lc.Errorf("invalid request topic: %s", err.Error())
					responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, err.Error())
					err = messageBus.Publish(responseEnvelope, internalResponseTopic)
					if err != nil {
						lc.Errorf("Could not publish to topic '%s': %s", internalResponseTopic, err.Error())
					}

					continue
				}

				err = validateGetCommandQueryParameters(requestEnvelope.QueryParams)
				if err != nil {
					lc.Errorf(err.Error())
					responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, err.Error())
					err = messageBus.Publish(responseEnvelope, internalResponseTopic)
					if err != nil {
						lc.Errorf("Could not publish to topic '%s': %s", internalResponseTopic, err.Error())
					}

					continue
				}

				// expected internal command request topic scheme: #/<device-service>/<device>/<command-name>/<method>
				err = messageBus.Publish(requestEnvelope, deviceRequestTopic)
				if err != nil {
					lc.Errorf("Could not publish to topic '%s': %s", deviceRequestTopic, err.Error())
					continue
				}

				lc.Debugf("Command request sent to internal MessageBus. Topic: %s, Correlation-id: %s", deviceRequestTopic, requestEnvelope.CorrelationID)
				router.SetResponseTopic(requestEnvelope.RequestID, internalResponseTopic, false)
			}
		}
	}()

	return nil
}

// SubscribeCommandQueryRequests subscribes command query requests from EdgeX service (e.g., Application Service)
// via internal MessageBus
func SubscribeCommandQueryRequests(ctx context.Context, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	messageBusInfo := container.ConfigurationFrom(dic.Get).MessageQueue
	internalQueryRequestTopic := messageBusInfo.Internal.Topics[QueryRequestTopic]
	internalQueryResponseTopic := messageBusInfo.Internal.Topics[QueryResponseTopic]

	messages := make(chan types.MessageEnvelope)
	messageErrors := make(chan error)
	topics := []types.TopicChannel{
		{
			Topic:    internalQueryRequestTopic,
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
				lc.Infof("Exiting waiting for MessageBus '%s' topic messages", internalQueryRequestTopic)
				return
			case err = <-messageErrors:
				lc.Error(err.Error())
			case requestEnvelope := <-messages:
				lc.Debugf("Command query request received on internal MessageBus. Topic: %s, Correlation-id: %s", requestEnvelope.ReceivedTopic, requestEnvelope.CorrelationID)

				// example topic scheme: /commandquery/request/<device>
				// deviceName is expected to be at last topic level.
				topicLevels := strings.Split(requestEnvelope.ReceivedTopic, "/")
				deviceName := topicLevels[len(topicLevels)-1]
				if strings.EqualFold(deviceName, common.All) {
					deviceName = common.All
				}

				responseEnvelope, err := getCommandQueryResponseEnvelope(requestEnvelope, deviceName, dic)
				if err != nil {
					lc.Error(err.Error())
					responseEnvelope = types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, err.Error())
				}

				err = messageBus.Publish(responseEnvelope, internalQueryResponseTopic)
				if err != nil {
					lc.Errorf("Could not publish to topic '%s': %s", internalQueryResponseTopic, err.Error())
					continue
				}

				lc.Debugf("Command query response sent to internal MessageBus. Topic: %s, Correlation-id: %s", internalQueryResponseTopic, requestEnvelope.CorrelationID)
			}
		}
	}()

	return nil
}
