//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	bootstrapMessaging "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-messaging/v2/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/fxamacker/cbor/v2"
)

// SubscribeEvents subscribes to events from message bus
func SubscribeEvents(ctx context.Context, dic *di.Container) errors.EdgeX {
	configuration := dataContainer.ConfigurationFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	messageBusInfo := configuration.MessageQueue

	messageBusInfo.AuthMode = strings.ToLower(strings.TrimSpace(messageBusInfo.AuthMode))
	if len(messageBusInfo.AuthMode) > 0 && messageBusInfo.AuthMode != bootstrapMessaging.AuthModeNone {
		if err := bootstrapMessaging.SetOptionsAuthData(&messageBusInfo, lc, dic); err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}

	messageBus, err := messaging.NewMessageClient(
		types.MessageBusConfig{
			SubscribeHost: types.HostInfo{
				Host:     messageBusInfo.Host,
				Port:     messageBusInfo.Port,
				Protocol: messageBusInfo.Protocol,
			},
			Type:     messageBusInfo.Type,
			Optional: messageBusInfo.Optional,
		})
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	err = messageBus.Connect()
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Infof("Subscribing to topic: '%s' @ %s://%s:%d",
		messageBusInfo.SubscribeTopic,
		messageBusInfo.Protocol,
		messageBusInfo.Host,
		messageBusInfo.Port)

	messages := make(chan types.MessageEnvelope)
	messageErrors := make(chan error)

	topics := []types.TopicChannel{
		{
			Topic:    messageBusInfo.SubscribeTopic,
			Messages: messages,
		},
	}

	err = messageBus.Subscribe(topics, messageErrors)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	go func() {
		for {
			select {
			case e := <-messageErrors:
				lc.Error(e.Error())
			case msgEnvelope := <-messages:
				lc.Debugf("Event received on message queue. Topic: %s, Correlation-id: %s ", messageBusInfo.SubscribeTopic, msgEnvelope.CorrelationID)
				event := &requests.AddEventRequest{}
				err = unmarshalPayload(msgEnvelope, event)
				if err != nil {
					lc.Errorf("fail to unmarshal event, %v", err)
					break
				}
				err = v2.Validate(event)
				if err != nil {
					lc.Errorf("invalid event, %v", err)
					break
				}
				err = AddEvent(requests.AddEventReqToEventModel(*event), ctx, dic)
				if err != nil {
					lc.Errorf("fail to persist the event, %v", err)
				}
			}
		}
	}()

	return nil
}

func unmarshalPayload(envelope types.MessageEnvelope, target interface{}) error {
	var err error
	switch envelope.ContentType {
	case clients.ContentTypeJSON:
		err = json.Unmarshal(envelope.Payload, target)

	case clients.ContentTypeCBOR:
		err = cbor.Unmarshal(envelope.Payload, target)

	default:
		err = fmt.Errorf("unsupported content-type '%s' recieved", envelope.ContentType)
	}
	return err
}
