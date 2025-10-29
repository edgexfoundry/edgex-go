//
// Copyright (C) 2020-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"
	"strings"

	msgTypes "github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/edgex-go/internal/core/data/query"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"

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
func (a *CoreDataApp) PublishEvent(data requestDTO.AddEventRequest, serviceName string, profileName string, deviceName string, sourceName string, ctx context.Context, dic *di.Container) {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	msgClient := bootstrapContainer.MessagingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)
	correlationId := correlation.FromContext(ctx)

	basePrefix := configuration.MessageBus.GetBaseTopicPrefix()
	publishTopic := common.NewPathBuilder().EnableNameFieldEscape(configuration.Service.EnableNameFieldEscape).
		SetPath(basePrefix).SetPath(common.EventsPublishTopic).SetPath(CoreDataEventTopicPrefix).SetNameFieldPath(serviceName).SetNameFieldPath(profileName).SetNameFieldPath(deviceName).SetNameFieldPath(sourceName).BuildPath()
	lc.Debugf("Publishing AddEventRequest to MessageBus. Topic: %s; %s: %s", publishTopic, common.CorrelationHeader, correlationId)

	msgEnvelope := msgTypes.NewMessageEnvelope(data, ctx)
	err := msgClient.PublishWithSizeLimit(msgEnvelope, publishTopic, configuration.MaxEventSize)
	if err != nil {
		lc.Errorf("Unable to send message for API event. Correlation-id: %s, Profile Name: %s, "+
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
	_, err := dbClient.EventById(id)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	err = dbClient.DeleteEventById(id)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return nil
}

// EventTotalCount return the count of all events currently stored in the database and error if any
func (a *CoreDataApp) EventTotalCount(dic *di.Container) (int64, errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)

	count, err := dbClient.EventTotalCount()
	if err != nil {
		return 0, errors.NewCommonEdgeXWrapper(err)
	}

	return count, nil
}

// EventCountByDeviceName return the count of all events associated with given device and error if any
func (a *CoreDataApp) EventCountByDeviceName(deviceName string, dic *di.Container) (int64, errors.EdgeX) {
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
func (a *CoreDataApp) AllEvents(parms query.Parameters, dic *di.Container) (events []dtos.Event, totalCount int64, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)

	eventModels, err := dbClient.AllEvents(parms.Offset, parms.Limit)
	if err != nil {
		return events, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	events = make([]dtos.Event, len(eventModels))
	for i, e := range eventModels {
		events[i] = dtos.FromEventModelToDTO(e)
		processNumericReadings(parms.Numeric, events[i].Readings)
	}
	if parms.Offset < 0 {
		return events, 0, err // skip total count
	}

	totalCount, err = dbClient.EventTotalCount()
	if err != nil {
		return events, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, parms.Offset, parms.Limit)
	if !cont {
		return []dtos.Event{}, totalCount, err
	}

	return events, totalCount, nil
}

// EventsByDeviceName query events with offset, limit and name
func (a *CoreDataApp) EventsByDeviceName(parms query.Parameters, name string, dic *di.Container) (events []dtos.Event, totalCount int64, err errors.EdgeX) {
	if name == "" {
		return events, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)

	eventModels, err := dbClient.EventsByDeviceName(parms.Offset, parms.Limit, name)
	if err != nil {
		return events, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	events = make([]dtos.Event, len(eventModels))
	for i, e := range eventModels {
		events[i] = dtos.FromEventModelToDTO(e)
		processNumericReadings(parms.Numeric, events[i].Readings)
	}
	if parms.Offset < 0 {
		return events, 0, err // skip total count
	}

	totalCount, err = dbClient.EventCountByDeviceName(name)
	if err != nil {
		return events, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, parms.Offset, parms.Limit)
	if !cont {
		return []dtos.Event{}, totalCount, err
	}
	return events, totalCount, nil
}

// EventsByTimeRange query events with offset, limit and time range
func (a *CoreDataApp) EventsByTimeRange(parms query.Parameters, dic *di.Container) (events []dtos.Event, totalCount int64, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)

	eventModels, err := dbClient.EventsByTimeRange(parms.Start, parms.End, parms.Offset, parms.Limit)
	if err != nil {
		return events, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	events = make([]dtos.Event, len(eventModels))
	for i, e := range eventModels {
		events[i] = dtos.FromEventModelToDTO(e)
		processNumericReadings(parms.Numeric, events[i].Readings)
	}
	if parms.Offset < 0 {
		return events, 0, err // skip total count
	}

	totalCount, err = dbClient.EventCountByTimeRange(parms.Start, parms.End)
	if err != nil {
		return events, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, parms.Offset, parms.Limit)
	if !cont {
		return []dtos.Event{}, totalCount, err
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
