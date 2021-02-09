//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	v2DataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/v2/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	dto "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
	msgTypes "github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/google/uuid"
)

// ValidateEvent validates if e is a valid event with corresponding device profile name and device name
// ValidateEvent throws error when profileName or deviceName doesn't match to e
func ValidateEvent(e models.Event, profileName string, deviceName string, ctx context.Context, dic *di.Container) errors.EdgeX {
	if e.ProfileName != profileName {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("event's profileName %s mismatches %s", e.ProfileName, profileName), nil)
	}
	if e.DeviceName != deviceName {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("event's deviceName %s mismatches %s", e.DeviceName, deviceName), nil)
	}
	return nil
}

// The AddEvent function accepts the new event model from the controller functions
// and invokes addEvent function in the infrastructure layer
func AddEvent(e models.Event, profileName string, deviceName string, ctx context.Context, dic *di.Container) (err errors.EdgeX) {
	configuration := dataContainer.ConfigurationFrom(dic.Get)
	if !configuration.Writable.PersistData {
		return nil
	}

	dbClient := v2DataContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	// Add the event and readings to the database
	if configuration.Writable.PersistData {
		correlationId := correlation.FromContext(ctx)
		addedEvent, err := dbClient.AddEvent(e)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
		e = addedEvent

		lc.Debug(fmt.Sprintf(
			"Event created on DB successfully. Event-id: %s, Correlation-id: %s ",
			e.Id,
			correlationId,
		))
	}

	return nil
}

// PublishEvent publishes incoming AddEventRequest through MessageClient
func PublishEvent(addEventReq dto.AddEventRequest, profileName string, deviceName string, ctx context.Context, dic *di.Container) {
	lc := container.LoggingClientFrom(dic.Get)
	msgClient := dataContainer.MessagingClientFrom(dic.Get)
	configuration := dataContainer.ConfigurationFrom(dic.Get)
	correlationId := correlation.FromContext(ctx)

	lc.Debug("Putting V2 Event DTO on message queue", clients.CorrelationHeader, correlationId)

	var data []byte
	var err error

	if len(clients.FromContext(ctx, clients.ContentType)) == 0 {
		ctx = context.WithValue(ctx, clients.ContentType, clients.ContentTypeJSON)
	}

	// Must make sure API Version for embedded DTOs is set since it isn't required by the request,
	// but is needed when published to Message Bus.
	addEventReq.Event.Versionable = common.NewVersionable()
	for index := range addEventReq.Event.Readings {
		addEventReq.Event.Readings[index].Versionable = common.NewVersionable()
	}

	data, err = json.Marshal(addEventReq)
	if err != nil {
		lc.Error(fmt.Sprintf("error marshaling V2 AddEventRequest DTO: %+v", addEventReq), clients.CorrelationHeader, correlationId)
		return
	}

	publishTopic := fmt.Sprintf("%s/%s/%s", configuration.MessageQueue.PublishTopicPrefix, profileName, deviceName)
	msgEnvelope := msgTypes.NewMessageEnvelope(data, ctx)
	err = msgClient.Publish(msgEnvelope, publishTopic)
	if err != nil {
		lc.Error(fmt.Sprintf("Unable to send message for V2 API event. Correlation-id: %s, Profile Name: %s, "+
			"Device Name: %s, Error: %v", correlationId, profileName, deviceName, err))
	} else {
		lc.Debug(fmt.Sprintf(
			"V2 API Event Published on message queue. Topic: %s, Correlation-id: %s ", publishTopic, correlationId))
	}
}

func EventById(id string, dic *di.Container) (dtos.Event, errors.EdgeX) {
	if id == "" {
		return dtos.Event{}, errors.NewCommonEdgeX(errors.KindInvalidId, "id is empty", nil)
	}
	_, err := uuid.Parse(id)
	if err != nil {
		return dtos.Event{}, errors.NewCommonEdgeX(errors.KindInvalidId, "fail to parse id as an UUID", err)
	}

	dbClient := v2DataContainer.DBClientFrom(dic.Get)

	event, err := dbClient.EventById(id)
	if err != nil {
		return dtos.Event{}, errors.NewCommonEdgeXWrapper(err)
	}

	eventDTO := dtos.FromEventModelToDTO(event)
	return eventDTO, nil
}

// The DeleteEventById function accepts event id from the controller functions
// and invokes DeleteEventById function in the infrastructure layer to remove
// event
func DeleteEventById(id string, dic *di.Container) errors.EdgeX {
	if id == "" {
		return errors.NewCommonEdgeX(errors.KindInvalidId, "id is empty", nil)
	} else {
		_, err := uuid.Parse(id)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindInvalidId, "Failed to parse ID as an UUID", err)
		}
	}

	dbClient := v2DataContainer.DBClientFrom(dic.Get)

	err := dbClient.DeleteEventById(id)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return nil
}

// EventTotalCount return the count of all of events currently stored in the database and error if any
func EventTotalCount(dic *di.Container) (uint32, errors.EdgeX) {
	dbClient := v2DataContainer.DBClientFrom(dic.Get)

	count, err := dbClient.EventTotalCount()
	if err != nil {
		return 0, errors.NewCommonEdgeXWrapper(err)
	}

	return count, nil
}

// EventCountByDeviceName return the count of all of events associated with given device and error if any
func EventCountByDeviceName(deviceName string, dic *di.Container) (uint32, errors.EdgeX) {
	dbClient := v2DataContainer.DBClientFrom(dic.Get)

	count, err := dbClient.EventCountByDeviceName(deviceName)
	if err != nil {
		return 0, errors.NewCommonEdgeXWrapper(err)
	}

	return count, nil
}

// The DeleteEventsByDeviceName function will be invoked by controller functions
// and then invokes DeleteEventsByDeviceName function in the infrastructure layer to remove
// all events/readings that are associated with the given deviceName
func DeleteEventsByDeviceName(deviceName string, dic *di.Container) errors.EdgeX {
	if len(strings.TrimSpace(deviceName)) <= 0 {
		return errors.NewCommonEdgeX(errors.KindInvalidId, "blank device name is not allowed", nil)
	}
	dbClient := v2DataContainer.DBClientFrom(dic.Get)

	err := dbClient.DeleteEventsByDeviceName(deviceName)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// AllEvents query events by offset and limit
func AllEvents(offset int, limit int, dic *di.Container) (events []dtos.Event, err errors.EdgeX) {
	dbClient := v2DataContainer.DBClientFrom(dic.Get)
	eventModels, err := dbClient.AllEvents(offset, limit)
	if err != nil {
		return events, errors.NewCommonEdgeXWrapper(err)
	}
	events = make([]dtos.Event, len(eventModels))
	for i, e := range eventModels {
		events[i] = dtos.FromEventModelToDTO(e)
	}
	return events, nil
}

// EventsByDeviceName query events with offset, limit and name
func EventsByDeviceName(offset int, limit int, name string, dic *di.Container) (events []dtos.Event, err errors.EdgeX) {
	if name == "" {
		return events, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := v2DataContainer.DBClientFrom(dic.Get)
	eventModels, err := dbClient.EventsByDeviceName(offset, limit, name)
	if err != nil {
		return events, errors.NewCommonEdgeXWrapper(err)
	}
	events = make([]dtos.Event, len(eventModels))
	for i, e := range eventModels {
		events[i] = dtos.FromEventModelToDTO(e)
	}
	return events, nil
}

// EventsByTimeRange query events with offset, limit and time range
func EventsByTimeRange(start int, end int, offset int, limit int, dic *di.Container) (events []dtos.Event, err errors.EdgeX) {
	dbClient := v2DataContainer.DBClientFrom(dic.Get)
	eventModels, err := dbClient.EventsByTimeRange(start, end, offset, limit)
	if err != nil {
		return events, errors.NewCommonEdgeXWrapper(err)
	}
	events = make([]dtos.Event, len(eventModels))
	for i, e := range eventModels {
		events[i] = dtos.FromEventModelToDTO(e)
	}
	return events, nil
}

// The DeleteEventsByAge function will be invoked by controller functions
// and then invokes DeleteEventsByAge function in the infrastructure layer to remove
// events that are older than age.  Age is supposed in milliseconds since created timestamp.
func DeleteEventsByAge(age int64, dic *di.Container) errors.EdgeX {
	dbClient := v2DataContainer.DBClientFrom(dic.Get)

	err := dbClient.DeleteEventsByAge(age)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}
