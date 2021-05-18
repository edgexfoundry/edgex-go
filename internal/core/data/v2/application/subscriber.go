//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"encoding/json"
	"fmt"

	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/fxamacker/cbor/v2"
)

// SubscribeEvents subscribes to events from message bus
func SubscribeEvents(ctx context.Context, dic *di.Container) errors.EdgeX {
	messageBusInfo := dataContainer.ConfigurationFrom(dic.Get).MessageQueue
	lc := container.LoggingClientFrom(dic.Get)

	messageBus := dataContainer.MessagingClientFrom(dic.Get)

	messages := make(chan types.MessageEnvelope)
	messageErrors := make(chan error)

	topics := []types.TopicChannel{
		{
			Topic:    messageBusInfo.SubscribeTopic,
			Messages: messages,
		},
	}

	err := messageBus.Subscribe(topics, messageErrors)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				lc.Infof("Exiting waiting for MessageBus '%s' topic messages", messageBusInfo.SubscribeTopic)
				return
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
