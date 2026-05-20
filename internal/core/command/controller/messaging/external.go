//
// Copyright (C) 2022-2026 IOTech Ltd
// Copyright (C) 2023 Intel Inc.
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"

	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"

	"github.com/edgexfoundry/edgex-go/internal/core/command/container"
)

// defaultMaxConcurrentExternalCommands caps in-flight external MQTT command requests so a single
// slow/offline device cannot block the Paho callback goroutine and starve subsequent commands.
// At capacity, new requests are rejected inline with a "service busy" response envelope —
// queuing further would just delay the same failure.
const defaultMaxConcurrentExternalCommands = 32

func OnConnectHandler(requestTimeout time.Duration, dic *di.Container) mqtt.OnConnectHandler {
	// Resolved once at service start; captured by the subscribe closure so it survives MQTT
	// reconnects. Hot-reload is intentionally not supported — resizing the channel mid-flight
	// would orphan in-flight work.
	maxInFlight := container.ConfigurationFrom(dic.Get).ExternalCommand.MaxConcurrentRequests
	if maxInFlight <= 0 {
		maxInFlight = defaultMaxConcurrentExternalCommands
	}
	sem := make(chan struct{}, maxInFlight)

	return func(client mqtt.Client) {
		lc := bootstrapContainer.LoggingClientFrom(dic.Get)
		config := container.ConfigurationFrom(dic.Get)
		externalTopics := config.ExternalMQTT.Topics
		qos := config.ExternalMQTT.QoS

		requestQueryTopic := externalTopics[common.CommandQueryRequestTopicKey]
		if token := client.Subscribe(requestQueryTopic, qos, commandQueryHandler(dic)); token.Wait() && token.Error() != nil {
			lc.Errorf("could not subscribe to topic '%s': %s", requestQueryTopic, token.Error().Error())
		} else {
			lc.Debugf("Subscribed to topic '%s' on external MQTT broker", requestQueryTopic)
		}

		requestCommandTopic := externalTopics[common.CommandRequestTopicKey]
		if token := client.Subscribe(requestCommandTopic, qos, commandRequestHandler(requestTimeout, sem, dic)); token.Wait() && token.Error() != nil {
			lc.Errorf("could not subscribe to topic '%s': %s", requestCommandTopic, token.Error().Error())
		} else {
			lc.Debugf("Subscribed to topic '%s' on external MQTT broker", requestCommandTopic)
		}
	}
}

func commandQueryHandler(dic *di.Container) mqtt.MessageHandler {
	return func(client mqtt.Client, message mqtt.Message) {
		lc := bootstrapContainer.LoggingClientFrom(dic.Get)
		lc.Debugf("Received command query request from external message broker on topic '%s' with %d bytes", message.Topic(), len(message.Payload()))

		requestEnvelope, err := types.NewMessageEnvelopeFromJSON(message.Payload())
		if err != nil {
			lc.Errorf("Failed to decode request MessageEnvelope: %s", err.Error())
			lc.Warn("Not publishing error message back due to insufficient information on response topic")
			return
		}

		externalMQTTInfo := container.ConfigurationFrom(dic.Get).ExternalMQTT
		responseTopic := externalMQTTInfo.Topics[common.CommandQueryResponseTopicKey]
		if responseTopic == "" {
			lc.Error("QueryResponseTopic not provided in External.Topics")
			lc.Warn("Not publishing error message back due to insufficient information on response topic")
			return
		}

		// example topic scheme: edgex/commandquery/request/<device-name>
		// deviceName is expected to be at last topic level.
		topicLevels := strings.Split(message.Topic(), "/")
		deviceName, err := url.PathUnescape(topicLevels[len(topicLevels)-1])
		if err != nil {
			lc.Errorf("Failed to unescape device name '%s': %s", deviceName, err.Error())
			lc.Warn("Not publishing error message back due to insufficient information on response topic")
			return
		}
		if strings.EqualFold(deviceName, common.All) {
			deviceName = common.All
		}

		responseEnvelope, err := getCommandQueryResponseEnvelope(requestEnvelope, deviceName, dic)
		if err != nil {
			responseEnvelope = types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, err.Error())
		}

		qos := externalMQTTInfo.QoS
		retain := externalMQTTInfo.Retain
		responseEnvelope.ReceivedTopic = responseTopic
		publishMessage(client, responseTopic, qos, retain, responseEnvelope, lc)
	}
}

func commandRequestHandler(requestTimeout time.Duration, sem chan struct{}, dic *di.Container) mqtt.MessageHandler {
	return func(client mqtt.Client, message mqtt.Message) {
		lc := bootstrapContainer.LoggingClientFrom(dic.Get)
		config := container.ConfigurationFrom(dic.Get)
		lc.Debugf("Received command request from external message broker on topic '%s' with %d bytes", message.Topic(), len(message.Payload()))

		externalMQTTInfo := container.ConfigurationFrom(dic.Get).ExternalMQTT
		qos := externalMQTTInfo.QoS
		retain := externalMQTTInfo.Retain

		requestEnvelope, err := types.NewMessageEnvelopeFromJSON(message.Payload())
		if err != nil {
			lc.Errorf("Failed to decode request MessageEnvelope: %s", err.Error())
			lc.Warn("Not publishing error message back due to insufficient information on response topic")
			return
		}

		topicLevels := strings.Split(message.Topic(), "/")
		length := len(topicLevels)
		if length < 3 {
			lc.Error("Failed to parse and construct response topic scheme, expected request topic scheme: '#/<device-name>/<command-name>/<method>")
			lc.Warn("Not publishing error message back due to insufficient information on response topic")
			return
		}

		// expected external command request/response topic scheme: #/<device-name>/<command-name>/<method>
		deviceName, err := url.PathUnescape(topicLevels[length-3])
		if err != nil {
			lc.Errorf("Failed to unescape device name from '%s': %s", topicLevels[length-3], err.Error())
			lc.Warn("Not publishing error message back due to insufficient information on response topic")
			return
		}
		commandName, err := url.PathUnescape(topicLevels[length-2])
		if err != nil {
			lc.Errorf("Failed to unescape command name from '%s': %s", topicLevels[length-2], err.Error())
			lc.Warn("Not publishing error message back due to insufficient information on response topic")
			return
		}
		method := topicLevels[length-1]
		if !strings.EqualFold(method, "get") && !strings.EqualFold(method, "set") {
			lc.Errorf("Unknown request method: %s, only 'get' or 'set' is allowed", method)
			lc.Warn("Not publishing error message back due to insufficient information on response topic")
			return
		}

		externalResponseTopic := common.BuildTopic(externalMQTTInfo.Topics[common.CommandResponseTopicPrefixKey], deviceName, commandName, method)

		internalBaseTopic := config.MessageBus.GetBaseTopicPrefix()
		topicPrefix := common.BuildTopic(internalBaseTopic, common.CoreCommandDeviceRequestPublishTopic)

		deviceServiceName, err := retrieveServiceNameByDevice(deviceName, dic)
		if err != nil {
			responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, err.Error())
			publishMessage(client, externalResponseTopic, qos, retain, responseEnvelope, lc)
			return
		}

		err = validateGetCommandQueryParameters(requestEnvelope.QueryParams)
		if err != nil {
			responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, err.Error())
			publishMessage(client, externalResponseTopic, qos, retain, responseEnvelope, lc)
			return
		}

		deviceRequestTopic := common.NewPathBuilder().EnableNameFieldEscape(config.Service.EnableNameFieldEscape).
			SetPath(topicPrefix).SetNameFieldPath(deviceServiceName).SetNameFieldPath(deviceName).SetNameFieldPath(commandName).SetPath(method).BuildPath()
		deviceResponseTopicPrefix := common.NewPathBuilder().EnableNameFieldEscape(config.Service.EnableNameFieldEscape).
			SetPath(internalBaseTopic).SetPath(common.ResponseTopic).SetNameFieldPath(deviceServiceName).BuildPath()

		lc.Debugf("Sending Command request to internal MessageBus. Topic: %s, Request-id: %s Correlation-id: %s", deviceRequestTopic, requestEnvelope.RequestID, requestEnvelope.CorrelationID)
		lc.Debugf("Expecting response on topic: %s/%s", deviceResponseTopicPrefix, requestEnvelope.RequestID)

		// Non-blocking acquire — never block the Paho callback goroutine. At capacity we reject
		// inline rather than queue, so a stuck device cannot cause head-of-line blocking on
		// subsequent external command requests.
		select {
		case sem <- struct{}{}:
		default:
			lc.Warnf("external command in-flight limit (%d) reached; rejecting request for device '%s'", cap(sem), deviceName)
			busy := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID,
				"core-command busy: too many concurrent external commands")
			publishMessage(client, externalResponseTopic, qos, retain, busy, lc)
			return
		}

		internalMessageBus := bootstrapContainer.MessagingClientFrom(dic.Get)

		go func() {
			defer func() { <-sem }()

			response, err := internalMessageBus.Request(requestEnvelope, deviceRequestTopic, deviceResponseTopicPrefix, requestTimeout)
			if err != nil {
				errorMessage := fmt.Sprintf("Failed to send DeviceCommand request with internal MessageBus: %v", err)
				publishMessage(client, externalResponseTopic, qos, retain,
					types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, errorMessage), lc)
				return
			}

			lc.Debugf("Command response received from internal MessageBus. Topic: %s, Request-id: %s Correlation-id: %s", response.ReceivedTopic, response.RequestID, response.CorrelationID)
			// Copy before mutating: concurrent goroutines must not share the same *MessageEnvelope.
			out := *response
			out.ReceivedTopic = externalResponseTopic
			publishMessage(client, externalResponseTopic, qos, retain, out, lc)
		}()
	}
}

func publishMessage(client mqtt.Client, responseTopic string, qos byte, retain bool, message types.MessageEnvelope, lc logger.LoggingClient) {
	if message.ErrorCode == 1 {
		lc.Errorf("%v", message.Payload)
	}

	envelopeBytes, _ := json.Marshal(&message)

	if token := client.Publish(responseTopic, qos, retain, envelopeBytes); token.Wait() && token.Error() != nil {
		lc.Errorf("Could not publish to external message broker on topic '%s': %s", responseTopic, token.Error())
	} else {
		lc.Debugf("Published response message to external message broker on topic '%s' with %d bytes", responseTopic, len(envelopeBytes))
	}
}
