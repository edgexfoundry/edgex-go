//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package external

import (
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
	RequestQueryTopic  = "RequestQueryTopic"
	ResponseQueryTopic = "ResponseQueryTopic"
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
