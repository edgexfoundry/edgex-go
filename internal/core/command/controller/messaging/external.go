//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/edgexfoundry/edgex-go/internal/core/command/application"
	"github.com/edgexfoundry/edgex-go/internal/core/command/container"
)

const (
	QueryRequestTopic          = "QueryRequestTopic"
	QueryResponseTopic         = "QueryResponseTopic"
	CommandRequestTopic        = "CommandRequestTopic"
	CommandResponseTopicPrefix = "CommandResponseTopicPrefix"
	DeviceRequestTopicPrefix   = "DeviceRequestTopicPrefix"
	DeviceResponseTopic        = "DeviceResponseTopic"
)

func OnConnectHandler(router MessagingRouter, dic *di.Container) mqtt.OnConnectHandler {
	return func(client mqtt.Client) {
		lc := bootstrapContainer.LoggingClientFrom(dic.Get)
		config := container.ConfigurationFrom(dic.Get)
		externalTopics := config.MessageQueue.External.Topics
		qos := config.MessageQueue.External.QoS
		retain := config.MessageQueue.External.Retain

		requestQueryTopic := externalTopics[QueryRequestTopic]
		responseQueryTopic := externalTopics[QueryResponseTopic]
		if token := client.Subscribe(requestQueryTopic, qos, commandQueryHandler(responseQueryTopic, qos, retain, dic)); token.Wait() && token.Error() != nil {
			lc.Errorf("could not subscribe to topic '%s': %s", responseQueryTopic, token.Error().Error())
			return
		}

		requestCommandTopic := externalTopics[CommandRequestTopic]
		if token := client.Subscribe(requestCommandTopic, qos, commandRequestHandler(router, dic)); token.Wait() && token.Error() != nil {
			lc.Errorf("could not subscribe to topic '%s': %s", responseQueryTopic, token.Error().Error())
			return
		}
	}
}

func commandQueryHandler(responseTopic string, qos byte, retain bool, dic *di.Container) mqtt.MessageHandler {
	return func(client mqtt.Client, message mqtt.Message) {
		var errorMessage string
		var responseEnvelope types.MessageEnvelope
		lc := bootstrapContainer.LoggingClientFrom(dic.Get)

		requestEnvelope, err := types.NewMessageEnvelopeFromJSON(message.Payload())
		if err != nil {
			lc.Errorf("Failed to decode request MessageEnvelope: %s", err.Error())
			lc.Warn("Not publishing error message back due to insufficient information on response topic")
			return
		}

		// example topic scheme: edgex/commandquery/request/<device>
		// deviceName is expected to be at last topic level.
		topicLevels := strings.Split(message.Topic(), "/")
		deviceName := topicLevels[len(topicLevels)-1]
		if strings.EqualFold(deviceName, common.All) {
			deviceName = common.All
		}

		var commands any
		var edgexErr edgexErr.EdgeX
		switch deviceName {
		case common.All:
			offset, limit := common.DefaultOffset, common.DefaultLimit
			if requestEnvelope.QueryParams != nil {
				if offsetRaw, ok := requestEnvelope.QueryParams[common.Offset]; ok {
					offset, err = strconv.Atoi(offsetRaw)
					if err != nil {
						errorMessage = fmt.Sprintf("Failed to convert 'offset' query parameter to intger: %s", err.Error())
						responseEnvelope = types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, errorMessage)
						publishMessage(client, responseTopic, qos, retain, responseEnvelope, lc)
						return
					}
				}
				if limitRaw, ok := requestEnvelope.QueryParams[common.Limit]; ok {
					limit, err = strconv.Atoi(limitRaw)
					if err != nil {
						errorMessage = fmt.Sprintf("Failed to convert 'limit' query parameter to integer: %s", err.Error())
						responseEnvelope = types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, errorMessage)
						publishMessage(client, responseTopic, qos, retain, responseEnvelope, lc)
						return
					}
				}
			}

			commands, _, edgexErr = application.AllCommands(offset, limit, dic)
			if edgexErr != nil {
				errorMessage = fmt.Sprintf("Failed to get all commands: %s", edgexErr.Error())
				responseEnvelope = types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, errorMessage)
				publishMessage(client, responseTopic, qos, retain, responseEnvelope, lc)
				return
			}
		default:
			commands, edgexErr = application.CommandsByDeviceName(deviceName, dic)
			if edgexErr != nil {
				errorMessage = fmt.Sprintf("Failed to get commands by device name '%s': %s", deviceName, edgexErr.Error())
				responseEnvelope = types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, errorMessage)
				publishMessage(client, responseTopic, qos, retain, responseEnvelope, lc)
				return
			}
		}

		payloadBytes, err := json.Marshal(commands)
		if err != nil {
			errorMessage = fmt.Sprintf("Failed to json encoding deviceCommands payload: %s", err.Error())
			responseEnvelope = types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, errorMessage)
			publishMessage(client, responseTopic, qos, retain, responseEnvelope, lc)
			return
		}

		responseEnvelope, _ = types.NewMessageEnvelopeForResponse(payloadBytes, requestEnvelope.RequestID, requestEnvelope.CorrelationID, common.ContentTypeJSON)
		publishMessage(client, responseTopic, qos, retain, responseEnvelope, lc)
	}
}

func commandRequestHandler(router MessagingRouter, dic *di.Container) mqtt.MessageHandler {
	return func(client mqtt.Client, message mqtt.Message) {
		lc := bootstrapContainer.LoggingClientFrom(dic.Get)
		messageBusInfo := container.ConfigurationFrom(dic.Get).MessageQueue
		qos := messageBusInfo.External.QoS
		retain := messageBusInfo.External.Retain

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
		deviceName := topicLevels[length-3]
		commandName := topicLevels[length-2]
		method := topicLevels[length-1]
		if !strings.EqualFold(method, "get") && !strings.EqualFold(method, "set") {
			lc.Errorf("Unknown request method: %s, only 'get' or 'set' is allowed", method)
			lc.Warn("Not publishing error message back due to insufficient information on response topic")
			return
		}
		externalResponseTopic := strings.Join([]string{messageBusInfo.External.Topics[CommandResponseTopicPrefix], deviceName, commandName, method}, "/")

		deviceRequestTopic, err := validateRequestTopic(messageBusInfo.Internal.Topics[DeviceRequestTopicPrefix], deviceName, commandName, method, dic)
		if err != nil {
			lc.Errorf("invalid request topic: %s", err.Error())
			responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, "nil Device Client")
			publishMessage(client, externalResponseTopic, qos, retain, responseEnvelope, lc)
			return
		}

		internalMessageBus := bootstrapContainer.MessagingClientFrom(dic.Get)
		err = internalMessageBus.Publish(requestEnvelope, deviceRequestTopic)
		if err != nil {
			errorMessage := fmt.Sprintf("Failed to send DeviceCommand request with internal MessageBus: %v", err)
			responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, errorMessage)
			publishMessage(client, externalResponseTopic, qos, retain, responseEnvelope, lc)
			return
		}

		lc.Debugf("Command request sent to internal MessageBus. Topic: %s, Correlation-id: %s", deviceRequestTopic, requestEnvelope.CorrelationID)
		router.SetResponseTopic(requestEnvelope.RequestID, externalResponseTopic, true)
	}
}

func publishMessage(client mqtt.Client, responseTopic string, qos byte, retain bool, message types.MessageEnvelope, lc logger.LoggingClient) {
	if message.ErrorCode == 1 {
		lc.Error(string(message.Payload))
	}

	envelopeBytes, _ := json.Marshal(&message)

	if token := client.Publish(responseTopic, qos, retain, envelopeBytes); token.Wait() && token.Error() != nil {
		lc.Errorf("Could not publish to external message queue on topic '%s': %s", responseTopic, token.Error())
	} else {
		lc.Debugf("Published response message to external message queue on topic '%s' with %d bytes", responseTopic, len(envelopeBytes))
	}
}

// validateRequestTopic validates the request topic by checking the existence of device and device service,
// returns the internal device request topic to which the command request will be sent.
func validateRequestTopic(prefix string, deviceName string, commandName string, method string, dic *di.Container) (string, error) {
	// retrieve device information through Metadata DeviceClient
	dc := bootstrapContainer.DeviceClientFrom(dic.Get)
	if dc == nil {
		return "", errors.New("nil Device Client")
	}
	deviceResponse, err := dc.DeviceByName(context.Background(), deviceName)
	if err != nil {
		return "", fmt.Errorf("failed to get Device by name %s: %v", deviceName, err)
	}

	// retrieve device service information through Metadata DeviceClient
	dsc := bootstrapContainer.DeviceServiceClientFrom(dic.Get)
	if dsc == nil {
		return "", errors.New("nil DeviceService Client")
	}
	deviceServiceResponse, err := dsc.DeviceServiceByName(context.Background(), deviceResponse.Device.ServiceName)
	if err != nil {
		return "", fmt.Errorf("failed to get DeviceService by name %s: %v", deviceResponse.Device.ServiceName, err)
	}

	// expected internal command request topic scheme: #/<device-service>/<device>/<command-name>/<method>
	return strings.Join([]string{prefix, deviceServiceResponse.Service.Name, deviceName, commandName, method}, "/"), nil

}
