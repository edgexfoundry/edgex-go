//
// Copyright (C) 2021-2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal/core/data/application"
	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/fxamacker/cbor/v2"
)

// SubscribeEvents subscribes to events from message bus
func SubscribeEvents(ctx context.Context, dic *di.Container) errors.EdgeX {
	messageBusInfo := dataContainer.ConfigurationFrom(dic.Get).MessageQueue
	lc := container.LoggingClientFrom(dic.Get)

	messageBus := container.MessagingClientFrom(dic.Get)

	messages := make(chan types.MessageEnvelope)
	messageErrors := make(chan error)

	app := application.CoreDataAppFrom(dic.Get)

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
				// decoding the large payload may cause memory issues so checking before decoding
				maxEventSize := dataContainer.ConfigurationFrom(dic.Get).MaxEventSize
				edgeXerr := utils.CheckPayloadSize(msgEnvelope.Payload, maxEventSize*1024)
				if edgeXerr != nil {
					lc.Errorf("event size exceed MaxEventSize(%d KB)", maxEventSize)
					break
				}
				err = unmarshalPayload(msgEnvelope, event)
				if err != nil {
					lc.Errorf("fail to unmarshal event, %v", err)
					break
				}
				err = validateEvent(msgEnvelope.ReceivedTopic, event.Event)
				if err != nil {
					lc.Error(err.Error())
					break
				}
				err = app.AddEvent(requests.AddEventReqToEventModel(*event), ctx, dic)
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
	case common.ContentTypeJSON:
		err = json.Unmarshal(envelope.Payload, target)

	case common.ContentTypeCBOR:
		err = cbor.Unmarshal(envelope.Payload, target)

	default:
		err = fmt.Errorf("unsupported content-type '%s' recieved", envelope.ContentType)
	}
	return err
}

func validateEvent(messageTopic string, e dtos.Event) errors.EdgeX {
	// Parse messageTopic by the pattern `edgex/events/<device-profile-name>/<device-name>/<source-name>`
	fields := strings.Split(messageTopic, "/")

	// assumes a non-empty base topic with /profileName/deviceName/sourceName appended by publisher
	if len(fields) < 4 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("invalid message topic %s", messageTopic), nil)
	}

	len := len(fields)
	profileName := fields[len-3]
	deviceName := fields[len-2]
	sourceName := fields[len-1]

	// Check whether the event fields match the message topic
	if e.ProfileName != profileName {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("event's profileName %s mismatches with the name %s received in topic", e.ProfileName, profileName), nil)
	}
	if e.DeviceName != deviceName {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("event's deviceName %s mismatches with the name %s received in topic", e.DeviceName, deviceName), nil)
	}
	if e.SourceName != sourceName {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("event's sourceName %s mismatches with the name %s received in topic", e.SourceName, sourceName), nil)
	}
	return nil
}
