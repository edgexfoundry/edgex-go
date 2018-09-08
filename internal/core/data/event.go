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
	"net/http"
	"net/url"
	"strconv"

	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
)

func countEvents() (int, error) {
	count, err := dbClient.EventCount()
	if err != nil {
		return -1, err
	}
	return count, nil
}

func countEventsByDevice(device string) (int, error) {
	err := newCheckDevice(device)
	if err != nil {
		return -1, err
	}

	count, err := dbClient.EventCountByDeviceId(device)
	if err != nil {
		return -1, fmt.Errorf("error obtaining count for device %s: %v", device, err)
	}
	return count, err
}

//TODO: Eliminate checkDevice below and make this checkDevice
func newCheckDevice(device string) error {
	if Configuration.MetaDataCheck {
		_, err := mdc.CheckForDevice(device)
		if err != nil {
			return err
		}
	}
	return nil
}

//TODO: Get rid of this method
// Check metadata if the device exists
func checkDevice(device string, w http.ResponseWriter) bool {
	if Configuration.MetaDataCheck {
		_, err := mdc.CheckForDevice(device)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("error checking device %s %v", device, err))
			switch err := err.(type) {
			case types.ErrNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
				return false
			default: //return an error on everything else.
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return false
			}
		}
	}

	return true
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

func getEvents(limit int) ([]models.Event, error) {
	var err error
	var events []models.Event

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

func addNew(e models.Event) (string, error) {
	retVal := "unsaved"
	err := newCheckDevice(e.Device)
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
		id, err := dbClient.AddEvent(&e)
		if err != nil {
			return "", err
		}
		retVal = id.Hex() //Coupling to Mongo in the model
	}

	putEventOnQueue(e)                              // Push the aux struct to export service (It has the actual readings)
	chEvents <- DeviceLastReported{e.Device}        // update last reported connected (device)
	chEvents <- DeviceServiceLastReported{e.Device} // update last reported connected (device service)

	return retVal, nil
}

func updateEvent(from models.Event) error {
	to, err := dbClient.EventById(from.ID.Hex())
	if err != nil {
		return errors.NewErrEventNotFound(from.ID.Hex())
	}

	// Update the fields
	if len(from.Device) > 0 {
		// Check device
		err = newCheckDevice(from.Device)
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
func deleteEvent(e models.Event) error {
	for _, reading := range e.Readings {
		if err := dbClient.DeleteReadingById(reading.Id.Hex()); err != nil {
			return err
		}
	}
	if err := dbClient.DeleteEventById(e.ID.Hex()); err != nil {
		return err
	}

	return nil
}

func deleteAllEvents() error {
	return dbClient.ScrubAllEvents()
}

func getEventById(id string) (models.Event, error) {
	e, err := dbClient.EventById(id)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrEventNotFound(id)
		}
		return models.Event{}, err
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

// Get event by device id
// Returns the events for the given device sorted by creation date and limited by 'limit'
// {deviceId} - the device that the events are for
// {limit} - the limit of events
// api/v1/event/device/{deviceId}/{limit}
func getEventByDeviceHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	limit := vars["limit"]
	deviceId, err := url.QueryUnescape(vars["deviceId"])

	// Problems unescaping URL
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error unescaping URL: " + err.Error())
		return
	}

	// Convert limit to int
	limitNum, err := strconv.Atoi(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error converting to integer: " + err.Error())
		return
	}

	// Check device
	if checkDevice(deviceId, w) == false {
		return
	}

	switch r.Method {
	case http.MethodGet:
		if limitNum > Configuration.ReadMaxLimit {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			LoggingClient.Error(maxExceededString)
			return
		}

		eventList, err := dbClient.EventsForDeviceLimit(deviceId, limitNum)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
			return
		}

		encode(eventList, w)
	}
}

// Delete all of the events associated with a device
// api/v1/event/device/{deviceId}
// 404 - device ID not found in metadata
// 503 - service unavailable
func deleteByDeviceIdHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	deviceId, err := url.QueryUnescape(vars["deviceId"])
	// Problems unescaping URL
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error unescaping the URL: " + err.Error())
		return
	}

	// Check device
	if checkDevice(deviceId, w) == false {
		return
	}

	switch r.Method {
	case http.MethodDelete:
		// Get the events by the device name
		events, err := dbClient.EventsForDevice(deviceId)
		if err != nil {
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		LoggingClient.Info("Deleting the events for device: " + deviceId)

		// Delete the events
		count := len(events)
		for _, event := range events {
			if err = deleteEvent(event); err != nil {
				LoggingClient.Error(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.Itoa(count)))
	}
}

// Get events by creation time
// {start} - start time, {end} - end time, {limit} - max number of results
// Sort the events by creation date
// 413 - number of results exceeds limit
// 503 - service unavailable
// api/v1/event/{start}/{end}/{limit}
func eventByCreationTimeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	start, err := strconv.ParseInt(vars["start"], 10, 64)
	// Problems converting start time
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Problem converting start time: " + err.Error())
		return
	}

	end, err := strconv.ParseInt(vars["end"], 10, 64)
	// Problems converting end time
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Problem converting end time: " + err.Error())
		return
	}

	limit, err := strconv.Atoi(vars["limit"])
	// Problems converting limit
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Problem converting limit: " + strconv.Itoa(limit))
		return
	}

	switch r.Method {
	case http.MethodGet:
		if limit > Configuration.ReadMaxLimit {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			LoggingClient.Error(maxExceededString)
			return
		}

		e, err := dbClient.EventsByCreationTime(start, end, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
			return
		}

		encode(e, w)
	}
}

// Get the readings for a device and filter them based on the value descriptor
// Only those readings whos name is the value descriptor should get through
// /event/device/{deviceId}/valuedescriptor/{valueDescriptor}/{limit}
// 413 - number exceeds limit
func readingByDeviceFilteredValueDescriptor(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	limit := vars["limit"]

	valueDescriptor, err := url.QueryUnescape(vars["valueDescriptor"])
	// Problems unescaping URL
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Problem unescaping value descriptor: " + err.Error())
		return
	}

	deviceId, err := url.QueryUnescape(vars["deviceId"])
	// Problems unescaping URL
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Problem unescaping device ID: " + err.Error())
		return
	}

	limitNum, err := strconv.Atoi(limit)
	// Problem converting the limit
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Problem converting limit to integer: " + err.Error())
		return
	}
	switch r.Method {
	case http.MethodGet:
		if limitNum > Configuration.ReadMaxLimit {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			LoggingClient.Error(maxExceededString)
			return
		}

		// Check device
		if checkDevice(deviceId, w) == false {
			return
		}

		// Get all the events for the device
		e, err := dbClient.EventsForDevice(deviceId)
		if err != nil {
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Only pick the readings who match the value descriptor
		readings := []models.Reading{}
		count := 0 // Make sure we stay below the limit
		for _, event := range e {
			if count >= limitNum {
				break
			}
			for _, reading := range event.Readings {
				if count >= limitNum {
					break
				}
				if reading.Name == valueDescriptor {
					readings = append(readings, reading)
					count += 1
				}
			}
		}

		encode(readings, w)
	}
}

// Scrub all the events that have been pushed
// Also remove the readings associated with the events
// api/v1/event/scrub
func scrubHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case http.MethodDelete:
		LoggingClient.Info("Scrubbing events.  Deleting all events that have been pushed")

		// Get the events
		events, err := dbClient.EventsPushed()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
			return
		}

		// Delete all the events
		count := len(events)
		for _, event := range events {
			if err = deleteEvent(event); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				LoggingClient.Error(err.Error())
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.Itoa(count)))
	}
}

// Put event on the message queue to be processed by the rules engine
func putEventOnQueue(e models.Event) {
	LoggingClient.Info("Putting event on message queue", "")
	//	Have multiple implementations (start with ZeroMQ)
	err := ep.SendEventMessage(e)
	if err != nil {
		LoggingClient.Error("Unable to send message for event: " + e.String())
	}
}
