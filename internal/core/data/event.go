/*******************************************************************************
 * Copyright 2017 Dell Inc.
 * Copyright (c) 2019 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/
package data

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/data/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation/models"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-messaging/messaging"
	msgTypes "github.com/edgexfoundry/go-mod-messaging/pkg/types"
)

const (
	ChecksumAlgoMD5    = "md5"
	ChecksumAlgoxxHash = "xxHash"
)

func countEvents(dbClient interfaces.DBClient) (int, error) {
	count, err := dbClient.EventCount()
	if err != nil {
		return -1, err
	}
	return count, nil
}

func countEventsByDevice(
	device string,
	ctx context.Context,
	dbClient interfaces.DBClient,
	mdc metadata.DeviceClient,
	configuration *config.ConfigurationStruct) (int, error) {

	err := checkDevice(device, ctx, mdc, configuration)
	if err != nil {
		return -1, err
	}

	count, err := dbClient.EventCountByDeviceId(device)
	if err != nil {
		return -1, fmt.Errorf("error obtaining count for device %s: %v", device, err)
	}
	return count, err
}

func deleteEventsByAge(age int64, lc logger.LoggingClient, dbClient interfaces.DBClient) (int, error) {
	events, err := dbClient.EventsOlderThanAge(age)
	if err != nil {
		return -1, err
	}

	// Delete all the events
	count := len(events)
	for _, event := range events {
		if err = deleteEvent(event, lc, dbClient); err != nil {
			return -1, err
		}
	}
	return count, nil
}

func getEvents(limit int, dbClient interfaces.DBClient) ([]contract.Event, error) {
	var err error
	var events []contract.Event

	if limit <= 0 {
		events, err = dbClient.Events()
	} else {
		events, err = dbClient.EventsWithLimit(limit)
	}

	if err != nil {
		return nil, err
	}
	return events, err
}

func addNewEvent(
	e models.Event, ctx context.Context,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	chEvents chan<- interface{},
	msgClient messaging.MessageClient,
	mdc metadata.DeviceClient,
	configuration *config.ConfigurationStruct) (string, error) {

	err := checkDevice(e.Device, ctx, mdc, configuration)
	if err != nil {
		return "", err
	}

	if configuration.Writable.ValidateCheck {
		lc.Debug("Validation enabled, parsing events")
		for reading := range e.Readings {
			// Check value descriptor
			name := e.Readings[reading].Name
			vd, err := dbClient.ValueDescriptorByName(name)
			if err != nil {
				if err == db.ErrNotFound {
					return "", errors.NewErrValueDescriptorNotFound(name)
				} else {
					return "", err
				}
			}
			err = isValidValueDescriptor(vd, e.Readings[reading])
			if err != nil {
				return "", err
			}
		}
	}

	// Add the event and readings to the database
	if configuration.Writable.PersistData {
		id, err := dbClient.AddEvent(e)
		if err != nil {
			return "", err
		}
		e.ID = id
	}

	putEventOnQueue(e, ctx, lc, msgClient, configuration) // Push event to message bus for App Services to consume
	chEvents <- DeviceLastReported{e.Device}              // update last reported connected (device)
	chEvents <- DeviceServiceLastReported{e.Device}       // update last reported connected (device service)

	return e.ID, nil
}

func updateEvent(
	from models.Event,
	ctx context.Context,
	dbClient interfaces.DBClient,
	mdc metadata.DeviceClient,
	configuration *config.ConfigurationStruct) error {

	to, err := dbClient.EventById(from.ID)
	if err != nil {
		return errors.NewErrEventNotFound(from.ID)
	}

	// Update the fields
	if len(from.Device) > 0 {
		// Check device
		err = checkDevice(from.Device, ctx, mdc, configuration)
		if err != nil {
			return err
		}

		// Set the device name on the event
		to.Device = from.Device
	}
	if from.Pushed != 0 {
		to.Pushed = from.Pushed
	}
	if from.Origin != 0 {
		to.Origin = from.Origin
	}

	mapped := models.Event{Event: to}
	return dbClient.UpdateEvent(mapped)
}

func deleteEventById(id string, lc logger.LoggingClient, dbClient interfaces.DBClient) error {
	e, err := getEventById(id, dbClient)
	if err != nil {
		return err
	}

	err = deleteEvent(e, lc, dbClient)
	if err != nil {
		return err
	}
	return nil
}

// Delete the event and readings
func deleteEvent(e contract.Event, lc logger.LoggingClient, dbClient interfaces.DBClient) error {
	for _, reading := range e.Readings {
		if err := deleteReadingById(reading.Id, lc, dbClient); err != nil {
			return err
		}
	}

	if err := dbClient.DeleteEventById(e.ID); err != nil {
		return err
	}

	return nil
}

func deleteAllEvents(dbClient interfaces.DBClient) error {
	return dbClient.ScrubAllEvents()
}

func getEventById(id string, dbClient interfaces.DBClient) (contract.Event, error) {
	e, err := dbClient.EventById(id)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrEventNotFound(id)
		}
		return contract.Event{}, err
	}
	return e, nil
}

// updateEventPushDateByChecksum updates the pushed dated for all events with a matching checksum which have not already been marked pushed
func updateEventPushDateByChecksum(
	checksum string,
	ctx context.Context,
	dbClient interfaces.DBClient,
	mdc metadata.DeviceClient,
	configuration *config.ConfigurationStruct) error {

	evts, err := dbClient.EventsByChecksum(checksum)
	if err != nil {
		return err
	}

	for _, e := range evts {
		e.Pushed = db.MakeTimestamp()
		// Updating the event has the desired side-effect of removing the checksum.
		// We only want the checksum for "marked pushed" functionality and once the event
		// has been marked pushed there is no reason to keep the checksum around.
		// The expectation is that above query will only return one result, but this is not guaranteed
		err = updateEvent(models.Event{Event: e}, ctx, dbClient, mdc, configuration)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateEventPushDate(
	id string,
	ctx context.Context,
	dbClient interfaces.DBClient,
	mdc metadata.DeviceClient,
	configuration *config.ConfigurationStruct) error {

	e, err := getEventById(id, dbClient)
	if err != nil {
		return err
	}

	e.Pushed = db.MakeTimestamp()
	err = updateEvent(models.Event{Event: e}, ctx, dbClient, mdc, configuration)
	if err != nil {
		return err
	}
	return nil
}

// Put event on the message queue to be processed by the rules engine
func putEventOnQueue(
	evt models.Event,
	ctx context.Context,
	lc logger.LoggingClient,
	msgClient messaging.MessageClient,
	configuration *config.ConfigurationStruct) {

	lc.Info("Putting event on message queue")

	evt.CorrelationId = correlation.FromContext(ctx)
	// Re-marshal JSON content into bytes.
	if clients.FromContext(clients.ContentType, ctx) == clients.ContentTypeJSON {
		data, err := json.Marshal(evt)
		if err != nil {
			lc.Error(fmt.Sprintf("error marshaling event: %s", evt.String()))
			return
		}
		evt.Bytes = data
	}

	msgEnvelope := msgTypes.NewMessageEnvelope(evt.Bytes, ctx)
	err := msgClient.Publish(msgEnvelope, configuration.MessageQueue.Topic)
	if err != nil {
		lc.Error(fmt.Sprintf("Unable to send message for event: %s %v", evt.String(), err))
	} else {
		lc.Info(fmt.Sprintf("Event Published on message queue. Topic: %s, Correlation-id: %s ", configuration.MessageQueue.Topic, msgEnvelope.CorrelationID))
	}
}

func getEventsByDeviceIdLimit(
	limit int,
	deviceId string,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) ([]contract.Event, error) {

	eventList, err := dbClient.EventsForDeviceLimit(deviceId, limit)
	if err != nil {
		lc.Error(err.Error())
		return nil, err
	}

	return eventList, nil
}

func getEventsByCreationTime(
	limit int,
	start int64,
	end int64,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) ([]contract.Event, error) {

	eventList, err := dbClient.EventsByCreationTime(start, end, limit)
	if err != nil {
		lc.Error(err.Error())
		return nil, err
	}

	return eventList, nil
}

func deleteEvents(deviceId string, dbClient interfaces.DBClient) (int, error) {
	return dbClient.DeleteEventsByDevice(deviceId)
}

func scrubPushedEvents(lc logger.LoggingClient, dbClient interfaces.DBClient) (int, error) {
	lc.Info("Scrubbing events.  Deleting all events that have been pushed")

	// Get the events
	events, err := dbClient.EventsPushed()
	if err != nil {
		lc.Error(err.Error())
		return 0, err
	}

	// Delete all the events
	count := len(events)
	for _, event := range events {
		if err = deleteEvent(event, lc, dbClient); err != nil {
			lc.Error(err.Error())
			return 0, err
		}
	}

	return count, nil
}
