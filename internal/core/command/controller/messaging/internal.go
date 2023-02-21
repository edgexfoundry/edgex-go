//
// Copyright (C) 2022-2023 IOTech Ltd
// Copyright (C) 2023 Intel Inc.
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-messaging/v3/messaging"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"

	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"

	"github.com/edgexfoundry/edgex-go/internal/core/command/container"
)

// SubscribeCommandRequests subscribes command requests from EdgeX service (e.g., Application Service)
// and forwards them to the appropriate Device Service via internal MessageBus
func SubscribeCommandRequests(ctx context.Context, requestTimeout time.Duration, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	baseTopic := container.ConfigurationFrom(dic.Get).MessageBus.GetBaseTopicPrefix()
	requestCommandTopic := common.BuildTopic(baseTopic, common.CoreCommandRequestSubscribeTopic)

	messages := make(chan types.MessageEnvelope)
	messageErrors := make(chan error)
	topics := []types.TopicChannel{
		{
			Topic:    requestCommandTopic,
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
				lc.Infof("Exiting waiting for MessageBus '%s' topic messages", requestCommandTopic)
				return
			case err = <-messageErrors:
				lc.Error(err.Error())
			case requestEnvelope := <-messages:
				processDeviceCommandRequest(messageBus, requestEnvelope, baseTopic, requestTimeout, lc, dic)
			}
		}
	}()

	return nil
}

func processDeviceCommandRequest(
	messageBus messaging.MessageClient,
	requestEnvelope types.MessageEnvelope,
	baseTopic string,
	requestTimeout time.Duration,
	lc logger.LoggingClient,
	dic *di.Container) {
	var err error

	lc.Debugf("Command device request received on internal MessageBus. Topic: %s, Request-id: %s, Correlation-id: %s", requestEnvelope.ReceivedTopic, requestEnvelope.RequestID, requestEnvelope.CorrelationID)

	if len(strings.TrimSpace(requestEnvelope.RequestID)) == 0 {
		lc.Errorf("RequestId not set in Command request received on internal MessageBus")
		lc.Warn("Not publishing error message back due to insufficient information to publish on response topic")
		return
	}

	// internal response topic scheme: <ResponseTopicPrefix>/<service-name>/<request-id>
	internalResponseTopic := common.BuildTopic(baseTopic, common.ResponseTopic, common.CoreCommandServiceKey, requestEnvelope.RequestID)
	topicLevels := strings.Split(requestEnvelope.ReceivedTopic, "/")
	length := len(topicLevels)
	if length < 3 {
		err = fmt.Errorf("invalid internal command request topic scheme. Expected request topic scheme with >=3 levels: '<DeviceRequestTopicPrefix>/<device-name>/<command-name>/<method>'")
		lc.Error(err.Error())
		responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, err.Error())
		err = messageBus.Publish(responseEnvelope, internalResponseTopic)
		if err != nil {
			lc.Errorf("Could not publish to topic '%s': %s", internalResponseTopic, err.Error())
		}
		return
	}

	// expected internal command request/response topic scheme: #/<device>/<command-name>/<method>
	deviceName := topicLevels[length-3]
	commandName, err := url.QueryUnescape(topicLevels[length-2])
	if err != nil {
		err = fmt.Errorf("failed to unescape command name '%s': %s", commandName, err.Error())
		lc.Error(err.Error())
		responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, err.Error())
		err = messageBus.Publish(responseEnvelope, internalResponseTopic)
		if err != nil {
			lc.Errorf("Could not publish to topic '%s': %s", internalResponseTopic, err.Error())
		}
		return
	}
	method := topicLevels[length-1]
	if !strings.EqualFold(method, "get") && !strings.EqualFold(method, "set") {
		err = fmt.Errorf("unknown request method: %s, only 'get' or 'set' is allowed", method)
		lc.Error(err.Error())
		responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, err.Error())
		err = messageBus.Publish(responseEnvelope, internalResponseTopic)
		if err != nil {
			lc.Errorf("Could not publish to topic '%s': %s", internalResponseTopic, err.Error())
		}
		return
	}

	topicPrefix := common.BuildTopic(baseTopic, common.CoreCommandDeviceRequestPublishTopic)
	// internal command request topic scheme: <DeviceRequestTopicPrefix>/<device-service>/<device>/<command-name>/<method>
	deviceServiceName, deviceRequestTopic, err := validateRequestTopic(topicPrefix, deviceName, commandName, method, dic)
	if err != nil {
		err = fmt.Errorf("invalid request topic: %s", err.Error())
		lc.Error(err.Error())
		responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, err.Error())
		err = messageBus.Publish(responseEnvelope, internalResponseTopic)
		if err != nil {
			lc.Errorf("Could not publish to topic '%s': %s", internalResponseTopic, err.Error())
		}
		return
	}

	err = validateGetCommandQueryParameters(requestEnvelope.QueryParams)
	if err != nil {
		lc.Errorf(err.Error())
		responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, err.Error())
		err = messageBus.Publish(responseEnvelope, internalResponseTopic)
		if err != nil {
			lc.Errorf("Could not publish to topic '%s': %s", internalResponseTopic, err.Error())
		}
		return
	}

	deviceResponseTopicPrefix := common.BuildTopic(baseTopic, common.ResponseTopic, deviceServiceName)

	lc.Debugf("Sending Command Device Request to internal MessageBus. Topic: %s, Correlation-id: %s", deviceRequestTopic, requestEnvelope.CorrelationID)
	lc.Debugf("Expecting response on topic: %s/%s", deviceResponseTopicPrefix, requestEnvelope.RequestID)

	response, err := messageBus.Request(requestEnvelope, deviceRequestTopic, deviceResponseTopicPrefix, requestTimeout)
	if err != nil {
		lc.Errorf("Request to topic '%s' failed: %s", deviceRequestTopic, err.Error())
		return
	}

	// original request is from internal MessageBus
	err = messageBus.Publish(*response, internalResponseTopic)
	if err != nil {
		lc.Errorf("Could not publish to internal MessageBus topic '%s': %s", internalResponseTopic, err.Error())
		return
	}

	lc.Debugf("Command response sent to internal MessageBus. Topic: %s, Correlation-id: %s", internalResponseTopic, response.CorrelationID)
}

// SubscribeCommandQueryRequests subscribes command query requests from EdgeX service (e.g., Application Service)
// via internal MessageBus
func SubscribeCommandQueryRequests(ctx context.Context, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	baseTopic := container.ConfigurationFrom(dic.Get).MessageBus.GetBaseTopicPrefix()
	queryRequestTopic := common.BuildTopic(baseTopic, common.CoreCommandQueryRequestSubscribeTopic)

	messages := make(chan types.MessageEnvelope)
	messageErrors := make(chan error)
	topics := []types.TopicChannel{
		{
			Topic:    queryRequestTopic,
			Messages: messages,
		},
	}

	messageBus := bootstrapContainer.MessagingClientFrom(dic.Get)

	lc.Infof("Subscribing to internal command query requests on topic: %s", queryRequestTopic)

	err := messageBus.Subscribe(topics, messageErrors)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				lc.Infof("Exiting waiting for MessageBus '%s' topic messages", queryRequestTopic)
				return
			case err = <-messageErrors:
				lc.Error(err.Error())
			case requestEnvelope := <-messages:
				processCommandQueryRequest(messageBus, requestEnvelope, baseTopic, lc, dic)
			}
		}
	}()

	return nil
}

func processCommandQueryRequest(
	messageBus messaging.MessageClient,
	requestEnvelope types.MessageEnvelope,
	baseTopic string,
	lc logger.LoggingClient,
	dic *di.Container,
) {
	lc.Debugf("Command query request received on internal MessageBus. Topic: %s, Request-id: %s, Correlation-id: %s", requestEnvelope.ReceivedTopic, requestEnvelope.RequestID, requestEnvelope.CorrelationID)

	if len(strings.TrimSpace(requestEnvelope.RequestID)) == 0 {
		lc.Errorf("RequestId not set in Command request received on internal MessageBus")
		lc.Warn("Not publishing error message back due to insufficient information to publish on response topic")
		return
	}

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

	// internal response topic scheme: <ResponseTopicPrefix>/<service-name>/<request-id>
	internalQueryResponseTopic := common.BuildTopic(baseTopic, common.ResponseTopic, common.CoreCommandServiceKey, requestEnvelope.RequestID)
	lc.Debugf("Responding to command query request on topic: %s", internalQueryResponseTopic)

	err = messageBus.Publish(responseEnvelope, internalQueryResponseTopic)
	if err != nil {
		lc.Errorf("Could not publish to topic '%s': %s", internalQueryResponseTopic, err.Error())
		return
	}

	lc.Debugf("Command query response sent to internal MessageBus. Topic: %s, Correlation-id: %s", internalQueryResponseTopic, requestEnvelope.CorrelationID)
}
