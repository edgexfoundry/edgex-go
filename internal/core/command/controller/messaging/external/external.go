//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package external

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/edgexfoundry/edgex-go/internal/core/command/application"
	"github.com/edgexfoundry/edgex-go/internal/core/command/container"
)

const (
	RequestQueryTopic          = "RequestQueryTopic"
	ResponseQueryTopic         = "ResponseQueryTopic"
	RequestCommandTopic        = "RequestCommandTopic"
	ResponseCommandTopicPrefix = "ResponseCommandTopicPrefix"
	RequestTopicPrefix         = "RequestTopicPrefix"
	ResponseTopic              = "ResponseTopic"
)

func OnConnectHandler(dic *di.Container) mqtt.OnConnectHandler {
	return func(client mqtt.Client) {
		lc := bootstrapContainer.LoggingClientFrom(dic.Get)
		config := container.ConfigurationFrom(dic.Get)
		externalTopics := config.MessageQueue.External.Topics
		qos := config.MessageQueue.External.QoS
		retain := config.MessageQueue.External.Retain

		requestQueryTopic := externalTopics[RequestQueryTopic]
		responseQueryTopic := externalTopics[ResponseQueryTopic]
		if token := client.Subscribe(requestQueryTopic, qos, commandQueryHandler(responseQueryTopic, qos, retain, dic)); token.Wait() && token.Error() != nil {
			lc.Errorf("could not subscribe to topic '%s': %s", responseQueryTopic, token.Error().Error())
			return
		}

		requestCommandTopic := externalTopics[RequestCommandTopic]
		if token := client.Subscribe(requestCommandTopic, qos, commandRequestHandler(dic)); token.Wait() && token.Error() != nil {
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
			responseEnvelope = types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, err.Error())
			publishMessage(client, responseTopic, qos, retain, responseEnvelope, lc)
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
		var edgexErr errors.EdgeX
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

func commandRequestHandler(dic *di.Container) mqtt.MessageHandler {
	return func(client mqtt.Client, message mqtt.Message) {
		lc := bootstrapContainer.LoggingClientFrom(dic.Get)
		messageBusInfo := container.ConfigurationFrom(dic.Get).MessageQueue
		qos := messageBusInfo.External.QoS
		retain := messageBusInfo.External.Retain

		// expected command request topic scheme: #/<device>/<command-name>/<method>
		topicLevels := strings.Split(message.Topic(), "/")
		length := len(topicLevels)
		deviceName := topicLevels[length-3]
		commandName := topicLevels[length-2]
		method := topicLevels[length-1]
		// expected command response topic scheme: #/<device>/<command-name>/<method>
		externalResponseTopic := strings.Join([]string{messageBusInfo.External.Topics[ResponseCommandTopicPrefix], deviceName, commandName, method}, "/")

		requestEnvelope, err := types.NewMessageEnvelopeFromJSON(message.Payload())
		if err != nil {
			responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, err.Error())
			publishMessage(client, externalResponseTopic, qos, retain, responseEnvelope, lc)
			return
		}

		if !strings.EqualFold(method, "get") && !strings.EqualFold(method, "set") {
			responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, fmt.Sprintf("unknown command method %s received", method))
			publishMessage(client, externalResponseTopic, qos, retain, responseEnvelope, lc)
			return
		}

		// retrieve device information through Metadata DeviceClient
		dc := bootstrapContainer.DeviceClientFrom(dic.Get)
		if dc == nil {
			responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, "nil Device Client")
			publishMessage(client, externalResponseTopic, qos, retain, responseEnvelope, lc)
			return
		}
		deviceResponse, err := dc.DeviceByName(context.Background(), deviceName)
		if err != nil {
			errorMessage := fmt.Sprintf("Failed to get Device by name %s: %v", deviceName, err)
			responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, errorMessage)
			publishMessage(client, externalResponseTopic, qos, retain, responseEnvelope, lc)
			return
		}

		// retrieve device service information through Metadata DeviceClient
		dsc := bootstrapContainer.DeviceServiceClientFrom(dic.Get)
		if dsc == nil {
			responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, "nil DeviceService Client")
			publishMessage(client, externalResponseTopic, qos, retain, responseEnvelope, lc)
			return
		}
		deviceServiceResponse, err := dsc.DeviceServiceByName(context.Background(), deviceResponse.Device.ServiceName)
		if err != nil {
			errorMessage := fmt.Sprintf("Failed to get DeviceService by name %s: %v", deviceResponse.Device.ServiceName, err)
			responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, errorMessage)
			publishMessage(client, externalResponseTopic, qos, retain, responseEnvelope, lc)
			return
		}

		// expected internal command request topic scheme: #/<device-service>/<device>/<command-name>/<method>
		internalRequestTopic := strings.Join([]string{messageBusInfo.Internal.Topics[RequestTopicPrefix], deviceServiceResponse.Service.Name, deviceName, commandName, method}, "/")
		internalMessageBus := bootstrapContainer.MessagingClientFrom(dic.Get)
		err = internalMessageBus.Publish(requestEnvelope, internalRequestTopic)
		if err != nil {
			errorMessage := fmt.Sprintf("Failed to send DeviceCommand request with internal MessageBus: %v", err)
			responseEnvelope := types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, errorMessage)
			publishMessage(client, externalResponseTopic, qos, retain, responseEnvelope, lc)
			return
		}
	}
}

func publishMessage(client mqtt.Client, responseTopic string, qos byte, retain bool, message types.MessageEnvelope, lc logger.LoggingClient) {
	if message.ErrorCode == 1 {
		lc.Error(string(message.Payload))
	}

	envelopeBytes, _ := json.Marshal(&message)

	if token := client.Publish(responseTopic, qos, retain, envelopeBytes); token.Wait() && token.Error() != nil {
		lc.Errorf("Could not publish to topic '%s': %s", responseTopic, token.Error())
	} else {
		lc.Debugf("Published response message on topic '%s' with %d bytes", responseTopic, len(envelopeBytes))
	}
}
