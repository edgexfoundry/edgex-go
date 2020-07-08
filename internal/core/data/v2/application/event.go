//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"encoding/json"
	"fmt"

	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	v2DataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/v2/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	msgTypes "github.com/edgexfoundry/go-mod-messaging/pkg/types"
)

// The AddEvent function accepts the new event model from the controller functions
// and invokes addEvent function in the infrastructure layer
func AddEvent(e models.Event, ctx context.Context, dic *di.Container) (string, error) {
	configuration := dataContainer.ConfigurationFrom(dic.Get)
	dbClient := v2DataContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	err := checkDevice(e.DeviceName, ctx, dic)
	if err != nil {
		return "", err
	}

	// Add the event and readings to the database
	if configuration.Writable.PersistData {
		correlationId := correlation.FromContext(ctx)
		e.CorrelationId = correlationId
		addedEvent, err := dbClient.AddEvent(e)
		if err != nil {
			return "", err
		}
		e = addedEvent

		lc.Info(fmt.Sprintf(
			"Event created on DB successfully. Event-id: %s, Correlation-id: %s ",
			e.Id,
			correlationId,
		))
	}

	//convert Event model to Event DTO
	eventDTO := dtos.FromEventModelToDTO(e)
	putEventOnQueue(eventDTO, ctx, dic) // Push event DTO to message bus for App Services to consume

	return e.Id, nil
}

// Put event DTO on the message queue to be processed by the rules engine
func putEventOnQueue(evt dtos.Event, ctx context.Context, dic *di.Container) {
	lc := container.LoggingClientFrom(dic.Get)
	msgClient := dataContainer.MessagingClientFrom(dic.Get)
	configuration := dataContainer.ConfigurationFrom(dic.Get)

	lc.Info("Putting V2 API event on message queue")

	var data []byte
	var err error
	// Re-marshal JSON content into bytes.
	if clients.FromContext(ctx, clients.ContentType) == clients.ContentTypeJSON {
		data, err = json.Marshal(evt)
		if err != nil {
			lc.Error(fmt.Sprintf("error marshaling event: %+v", evt))
			return
		}
	}

	msgEnvelope := msgTypes.NewMessageEnvelope(data, ctx)
	err = msgClient.Publish(msgEnvelope, configuration.MessageQueue.Topic)
	if err != nil {
		lc.Error(fmt.Sprintf("Unable to send message for V2 API event. Correlation-id: %s, Device Name: %s, Error: %v", correlation.FromContext(ctx), evt.DeviceName, err))
	} else {
		lc.Info(fmt.Sprintf(
			"Event Published on message queue. Topic: %s, Correlation-id: %s ",
			configuration.MessageQueue.Topic,
			msgEnvelope.CorrelationID,
		))
	}
}
