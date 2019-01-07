/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
)

func countEvents() (int, error) {
	count, err := dbClient.EventCount()
	if err != nil {
		return -1, err
	}
	return count, nil
}

func countEventsByDevice(device string) (int, error) {
	err := checkDevice(device)
	if err != nil {
		return -1, err
	}

	count, err := dbClient.EventCountByDeviceId(device)
	if err != nil {
		return -1, fmt.Errorf("error obtaining count for device %s: %v", device, err)
	}
	return count, err
}

func deleteEventsByAge(age int64) (int, error) {
	events, err := dbClient.EventsOlderThanAge(age)
	if err != nil {
		return -1, err
	}

	// Delete all the events
	count := len(events)
	for _, event := range events {
		if err = deleteEvent(event); err != nil {
			return -1, err
		}
	}
	return count, nil
}

func getEvents(limit int) ([]contract.Event, error) {
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

func addNewEvent(e contract.Event) (string, error) {
	err := checkDevice(e.Device)
	if err != nil {
		return "", err
	}

	if Configuration.ValidateCheck {
		LoggingClient.Debug("Validation enabled, parsing events")
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
	if Configuration.PersistData {
		id, err := dbClient.AddEvent(e)
		if err != nil {
			return "", err
		}
		e.ID = id
	}

	putEventOnQueue(e)                              // Push the aux struct to export service (It has the actual readings)
	chEvents <- DeviceLastReported{e.Device}        // update last reported connected (device)
	chEvents <- DeviceServiceLastReported{e.Device} // update last reported connected (device service)

	return e.ID, nil
}

func updateEvent(from contract.Event) error {
	to, err := dbClient.EventById(from.ID)
	if err != nil {
		return errors.NewErrEventNotFound(from.ID)
	}

	// Update the fields
	if len(from.Device) > 0 {
		// Check device
		err = checkDevice(from.Device)
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
	return dbClient.UpdateEvent(to)
}

func deleteEventById(id string) error {
	e, err := getEventById(id)
	if err != nil {
		return err
	}

	err = deleteEvent(e)
	if err != nil {
		return err
	}
	return nil
}

// Delete the event and readings
func deleteEvent(e contract.Event) error {
	for _, reading := range e.Readings {
		if err := deleteReadingById(reading.Id); err != nil {
			return err
		}
	}
	if err := dbClient.DeleteEventById(e.ID); err != nil {
		return err
	}

	return nil
}

func deleteAllEvents() error {
	return dbClient.ScrubAllEvents()
}

func getEventById(id string) (contract.Event, error) {
	e, err := dbClient.EventById(id)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrEventNotFound(id)
		}
		return contract.Event{}, err
	}
	return e, nil
}

func updateEventPushDate(id string) error {
	e, err := getEventById(id)
	if err != nil {
		return err
	}

	e.Pushed = db.MakeTimestamp()
	err = updateEvent(e)
	if err != nil {
		return err
	}
	return nil
}

// Put event on the message queue to be processed by the rules engine
func putEventOnQueue(e contract.Event) {
	LoggingClient.Info("Putting event on message queue")
	//	Have multiple implementations (start with ZeroMQ)
	err := ep.SendEventMessage(e)
	if err != nil {
		LoggingClient.Error("Unable to send message for event: " + e.String())
	}
}

func getEventsByDeviceIdLimit(limit int, deviceId string) ([]contract.Event, error) {
	eventList, err := dbClient.EventsForDeviceLimit(deviceId, limit)
	if err != nil {
		LoggingClient.Error(err.Error())
		return nil, err
	}

	return eventList, nil
}

func getEventsByCreationTime(limit int, start int64, end int64) ([]contract.Event, error) {
	eventList, err := dbClient.EventsByCreationTime(start, end, limit)
	if err != nil {
		LoggingClient.Error(err.Error())
		return nil, err
	}

	return eventList, nil
}

func deleteEvents(deviceId string) (int, error) {
	// Get the events by the device name
	events, err := dbClient.EventsForDevice(deviceId)
	if err != nil {
		LoggingClient.Error(err.Error())
		return 0, err
	}

	LoggingClient.Info("Deleting the events for device: " + deviceId)

	// Delete the events
	count := len(events)
	for _, event := range events {
		if err = deleteEvent(event); err != nil {
			LoggingClient.Error(err.Error())
			return 0, err
		}
	}

	return count, nil
}

func scrubPushedEvents() (int, error) {
	LoggingClient.Info("Scrubbing events.  Deleting all events that have been pushed")

	// Get the events
	events, err := dbClient.EventsPushed()
	if err != nil {
		LoggingClient.Error(err.Error())
		return 0, err
	}

	// Delete all the events
	count := len(events)
	for _, event := range events {
		if err = deleteEvent(event); err != nil {
			LoggingClient.Error(err.Error())
			return 0, err
		}
	}

	return count, nil
}
