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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
)

// Check metadata if the device exists
func checkDevice(device string, w http.ResponseWriter) bool {
	if configuration.MetaDataCheck {
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

// Put event on the message queue to be processed by the rules engine
func putEventOnQueue(e models.Event) {
	LoggingClient.Info("Putting event on message queue", "")
	//	Have multiple implementations (start with ZeroMQ)
	err := ep.SendEventMessage(e)
	if err != nil {
		LoggingClient.Error("Unable to send message for event: " + e.String())
	}
}

// Update when the device was last reported connected
func updateDeviceLastReportedConnected(device string) {
	// Config set to skip update last reported
	if !configuration.DeviceUpdateLastConnected {
		LoggingClient.Debug("Skipping update of device connected/reported times for:  " + device)
		return
	}

	d, err := mdc.CheckForDevice(device)
	if err != nil {
		LoggingClient.Error("Error getting device " + device + ": " + err.Error())
		return
	}

	// Couldn't find device
	if len(d.Name) == 0 {
		LoggingClient.Error("Error updating device connected/reported times.  Unknown device with identifier of:  " + device)
		return
	}

	t := time.Now().UnixNano() / int64(time.Millisecond)
	// Found device, now update lastReported
	err = mdc.UpdateLastConnectedByName(d.Name, t)
	if err != nil {
		LoggingClient.Error("Problems updating last connected value for device: " + d.Name)
		return
	}
	err = mdc.UpdateLastReportedByName(d.Name, t)
	if err != nil {
		LoggingClient.Error("Problems updating last reported value for device: " + d.Name)
	}
	return
}

// Update when the device service was last reported connected
func updateDeviceServiceLastReportedConnected(device string) {
	if !configuration.ServiceUpdateLastConnected {
		LoggingClient.Debug("Skipping update of device service connected/reported times for:  " + device)
		return
	}

	t := time.Now().UnixNano() / int64(time.Millisecond)

	// Get the device
	d, err := mdc.CheckForDevice(device)
	if err != nil {
		LoggingClient.Error("Error getting device " + device + ": " + err.Error())
		return
	}

	// Couldn't find device
	if len(d.Name) == 0 {
		LoggingClient.Error("Error updating device connected/reported times.  Unknown device with identifier of:  " + device)
		return
	}

	// Get the device service
	s := d.Service
	if &s == nil {
		LoggingClient.Error("Error updating device service connected/reported times.  Unknown device service in device:  " + d.Name)
		return
	}

	msc.UpdateLastConnected(s.Service.Id.Hex(), t)
	msc.UpdateLastReported(s.Service.Id.Hex(), t)
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

// Undocumented feature to remove all readings and events from the database
// This should primarily be used for debugging purposes
func scrubAllHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case http.MethodDelete:
		LoggingClient.Info("Deleting all events from database")

		err := dbClient.ScrubAllEvents()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			LoggingClient.Error("Error scrubbing all events/readings: " + err.Error())
			return
		}

		encode(true, w)
	}
}

/*
Handler for the event API
Status code 404 - event not found
Status code 413 - number of events exceeds limit
Status code 503 - unanticipated issues
api/v1/event
*/
func eventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	switch r.Method {
	// Get all events
	case http.MethodGet:
		events, err := dbClient.Events()
		if err != nil {
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Check max limit
		if len(events) > configuration.ReadMaxLimit {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			LoggingClient.Error(maxExceededString)
			return
		}

		encode(events, w)
		break
	// Post a new event
	case http.MethodPost:
		var e models.Event
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&e)

		// Problem Decoding Event
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			LoggingClient.Error("Error decoding event: " + err.Error())
			return
		}

		LoggingClient.Info("Posting Event: " + e.String())

		// Check device
		if checkDevice(e.Device, w) == false {
			return
		}

		if configuration.ValidateCheck {
			LoggingClient.Debug("Validation enabled, parsing events")
			for reading := range e.Readings {
				// Check value descriptor
				vd, err := dbClient.ValueDescriptorByName(e.Readings[reading].Name)
				if err != nil {
					if err == db.ErrNotFound {
						http.Error(w, "Value descriptor for a reading not found", http.StatusNotFound)
					} else {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
					LoggingClient.Error(err.Error())
					return
				}

				valid, err := isValidValueDescriptor(vd, e.Readings[reading])
				if !valid {
					http.Error(w, "Validation failed", http.StatusConflict)
					LoggingClient.Error("Validation failed")
					return
				}
			}
		}

		// Add the event and readings to the database
		if configuration.PersistData {
			id, err := dbClient.AddEvent(&e)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				LoggingClient.Error(err.Error())
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(id.Hex()))
		} else {
			encode("unsaved", w)
		}

		putEventOnQueue(e)                                 // Push the aux struct to export service (It has the actual readings)
		updateDeviceLastReportedConnected(e.Device)        // update last reported connected (device)
		updateDeviceServiceLastReportedConnected(e.Device) // update last reported connected (device service)

		break
	// Do not update the readings
	case http.MethodPut:
		var from models.Event
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&from)

		// Problem decoding event
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			LoggingClient.Error("Error decoding the event: " + err.Error())
			return
		}

		// Check if the event exists
		to, err := dbClient.EventById(from.ID.Hex())
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Event not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}

		LoggingClient.Info("Updating event: " + from.ID.Hex())

		// Update the fields
		if len(from.Device) > 0 {
			// Check device
			if checkDevice(from.Device, w) == false {
				return
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

		// Update
		if err = dbClient.UpdateEvent(to); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
	}
}

//GET
//Return the event specified by the event ID
///api/v1/event/{id}
//id - ID of the event to return
func getEventByIdHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	switch r.Method {
	case http.MethodGet:
		// URL parameters
		vars := mux.Vars(r)
		id := vars["id"]

		// Get the event
		e, err := dbClient.EventById(id)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Event not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}

		// Return the result
		encode(e, w)
	}
}

/*
Return number of events in Core Data
/api/v1/event/count
*/
func eventCountHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case http.MethodGet:
		count, err := dbClient.EventCount()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error(), "")
			return
		}

		// Return result
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(strconv.Itoa(count)))
		if err != nil {
			LoggingClient.Error(err.Error(), "")
		}
	}
}

/*
Return number of events for a given device in Core Data
deviceID - ID of the device to get count for
/api/v1/event/count/{deviceId}
*/
func eventCountByDeviceIdHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	id, err := url.QueryUnescape(vars["deviceId"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Problem unescaping URL: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Check device
		if checkDevice(id, w) == false {
			return
		}

		count, err := dbClient.EventCountByDeviceId(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
			return
		}

		// Return result
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.Itoa(count)))
		break
	}
}

/*
DELETE, PUT
Handle events specified by an ID
/api/v1/event/id/{id}
404 - ID not found
*/
func eventIdHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	id := vars["id"]

	switch r.Method {
	// Set the 'pushed' timestamp for the event to the current time - event is going to another (not fuse) service
	case http.MethodPut:
		// Check if the event exists
		e, err := dbClient.EventById(id)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Event not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}

		LoggingClient.Info("Updating event: " + e.ID.Hex())

		e.Pushed = time.Now().UnixNano() / int64(time.Millisecond)
		err = dbClient.UpdateEvent(e)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
		break
	// Delete the event and all of it's readings
	case http.MethodDelete:
		// Check if the event exists
		e, err := dbClient.EventById(id)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Event not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}

		LoggingClient.Info("Deleting event: " + e.ID.Hex())

		if err = deleteEvent(e); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
	}
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
		if limitNum > configuration.ReadMaxLimit {
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
		if limit > configuration.ReadMaxLimit {
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
		if limitNum > configuration.ReadMaxLimit {
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

// Remove all the old events and associated readings (by age)
// event/removeold/age/{age}
func eventByAgeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	age, err := strconv.ParseInt(vars["age"], 10, 64)

	// Problem converting age
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error converting the age to an integer")
		return
	}

	switch r.Method {
	case http.MethodDelete:
		// Get the events
		events, err := dbClient.EventsOlderThanAge(age)
		if err != nil {
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Delete all the events
		count := len(events)
		for _, event := range events {
			if err = deleteEvent(event); err != nil {
				LoggingClient.Error(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		LoggingClient.Info("Deleting events by age: " + vars["age"])

		// Return the count
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.Itoa(count)))
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
