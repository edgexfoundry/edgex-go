//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	msgTypes "github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"

	"github.com/google/uuid"
)

const CoreDataEventTopicPrefix = "core"

// ValidateEvent validates if e is a valid event with corresponding device profile name and device name and source name
// ValidateEvent throws error when profileName or deviceName doesn't match to e
func (a *CoreDataApp) ValidateEvent(e models.Event, profileName string, deviceName string, sourceName string, _ context.Context, _ *di.Container) errors.EdgeX {
	if e.ProfileName != profileName {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("event's profileName %s mismatches %s", e.ProfileName, profileName), nil)
	}
	if e.DeviceName != deviceName {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("event's deviceName %s mismatches %s", e.DeviceName, deviceName), nil)
	}
	if e.SourceName != sourceName {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("event's sourceName %s mismatches %s", e.SourceName, sourceName), nil)
	}
	return nil
}

// The AddEvent function accepts the new event model from the controller functions
// and invokes addEvent function in the infrastructure layer
func (a *CoreDataApp) AddEvent(e models.Event, ctx context.Context, dic *di.Container) (err errors.EdgeX) {
	configuration := container.ConfigurationFrom(dic.Get)
	if !configuration.Writable.PersistData {
		return nil
	}

	dbClient := container.DBClientFrom(dic.Get)

	// Add the event and readings to the database
	if configuration.Writable.PersistData {
		correlationId := correlation.FromContext(ctx)
		addedEvent, err := dbClient.AddEvent(e)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
		e = addedEvent

		a.lc.Debugf(
			"Event created on DB successfully. Event-id: %s, Correlation-id: %s ",
			e.Id,
			correlationId,
		)

		a.eventsPersistedCounter.Inc(1)
		a.readingsPersistedCounter.Inc(int64(len(addedEvent.Readings)))
	}

	return nil
}

// PublishEvent publishes incoming AddEventRequest in the format of []byte through MessageClient
func (a *CoreDataApp) PublishEvent(data []byte, serviceName string, profileName string, deviceName string, sourceName string, ctx context.Context, dic *di.Container) {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	msgClient := bootstrapContainer.MessagingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)
	correlationId := correlation.FromContext(ctx)

	basePrefix := configuration.MessageBus.GetBaseTopicPrefix()
	publishTopic := common.BuildTopic(basePrefix, common.EventsPublishTopic, CoreDataEventTopicPrefix, serviceName, profileName, deviceName, url.QueryEscape(sourceName))
	lc.Debugf("Publishing AddEventRequest to MessageBus. Topic: %s; %s: %s", publishTopic, common.CorrelationHeader, correlationId)

	msgEnvelope := msgTypes.NewMessageEnvelope(data, ctx)
	err := msgClient.Publish(msgEnvelope, publishTopic)
	if err != nil {
		lc.Errorf("Unable to send message for V2 API event. Correlation-id: %s, Profile Name: %s, "+
			"Device Name: %s, Source Name: %s, Error: %v", correlationId, profileName, deviceName, sourceName, err)
	} else {
		lc.Debugf("Event Published to MessageBus. Topic: %s, Correlation-id: %s ", publishTopic, correlationId)
	}
}

func (a *CoreDataApp) EventById(id string, dic *di.Container) (dtos.Event, errors.EdgeX) {
	if id == "" {
		return dtos.Event{}, errors.NewCommonEdgeX(errors.KindInvalidId, "id is empty", nil)
	}
	_, err := uuid.Parse(id)
	if err != nil {
		return dtos.Event{}, errors.NewCommonEdgeX(errors.KindInvalidId, "fail to parse id as an UUID", err)
	}

	dbClient := container.DBClientFrom(dic.Get)

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
func (a *CoreDataApp) DeleteEventById(id string, dic *di.Container) errors.EdgeX {
	if id == "" {
		return errors.NewCommonEdgeX(errors.KindInvalidId, "id is empty", nil)
	} else {
		_, err := uuid.Parse(id)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindInvalidId, "Failed to parse ID as an UUID", err)
		}
	}

	dbClient := container.DBClientFrom(dic.Get)

	err := dbClient.DeleteEventById(id)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return nil
}

// EventTotalCount return the count of all events currently stored in the database and error if any
func (a *CoreDataApp) EventTotalCount(dic *di.Container) (uint32, errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)

	count, err := dbClient.EventTotalCount()
	if err != nil {
		return 0, errors.NewCommonEdgeXWrapper(err)
	}

	return count, nil
}

// EventCountByDeviceName return the count of all events associated with given device and error if any
func (a *CoreDataApp) EventCountByDeviceName(deviceName string, dic *di.Container) (uint32, errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)

	count, err := dbClient.EventCountByDeviceName(deviceName)
	if err != nil {
		return 0, errors.NewCommonEdgeXWrapper(err)
	}

	return count, nil
}

// The DeleteEventsByDeviceName function will be invoked by controller functions
// and then invokes DeleteEventsByDeviceName function in the infrastructure layer to remove
// all events/readings that are associated with the given deviceName
func (a *CoreDataApp) DeleteEventsByDeviceName(deviceName string, dic *di.Container) errors.EdgeX {
	if len(strings.TrimSpace(deviceName)) <= 0 {
		return errors.NewCommonEdgeX(errors.KindInvalidId, "blank device name is not allowed", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	go func() {
		err := dbClient.DeleteEventsByDeviceName(deviceName)
		if err != nil {
			lc.Errorf("Delete events by device name failed: %v", err)
		}
	}()
	return nil
}

// AllEvents query events by offset and limit
func (a *CoreDataApp) AllEvents(offset int, limit int, dic *di.Container) (events []dtos.Event, totalCount uint32, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	eventModels, err := dbClient.AllEvents(offset, limit)
	if err == nil {
		totalCount, err = dbClient.EventTotalCount()
	}
	if err != nil {
		return events, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	events = make([]dtos.Event, len(eventModels))
	for i, e := range eventModels {
		events[i] = dtos.FromEventModelToDTO(e)
	}
	return events, totalCount, nil
}

// EventsByDeviceName query events with offset, limit and name
func (a *CoreDataApp) EventsByDeviceName(offset int, limit int, name string, dic *di.Container) (events []dtos.Event, totalCount uint32, err errors.EdgeX) {
	if name == "" {
		return events, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	eventModels, err := dbClient.EventsByDeviceName(offset, limit, name)
	if err == nil {
		totalCount, err = dbClient.EventCountByDeviceName(name)
	}
	if err != nil {
		return events, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	events = make([]dtos.Event, len(eventModels))
	for i, e := range eventModels {
		events[i] = dtos.FromEventModelToDTO(e)
	}
	return events, totalCount, nil
}

// EventsByTimeRange query events with offset, limit and time range
func (a *CoreDataApp) EventsByTimeRange(startTime int, endTime int, offset int, limit int, dic *di.Container) (events []dtos.Event, totalCount uint32, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	eventModels, err := dbClient.EventsByTimeRange(startTime, endTime, offset, limit)
	if err == nil {
		totalCount, err = dbClient.EventCountByTimeRange(startTime, endTime)
	}
	if err != nil {
		return events, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	events = make([]dtos.Event, len(eventModels))
	for i, e := range eventModels {
		events[i] = dtos.FromEventModelToDTO(e)
	}
	return events, totalCount, nil
}

// The DeleteEventsByAge function will be invoked by controller functions
// and then invokes DeleteEventsByAge function in the infrastructure layer to remove
// events that are older than age.  Age is supposed in milliseconds since created timestamp.
func (a *CoreDataApp) DeleteEventsByAge(age int64, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	go func() {
		err := dbClient.DeleteEventsByAge(age)
		if err != nil {
			lc.Errorf("Delete events by age failed: %v", err)
		}
	}()
	return nil
}
