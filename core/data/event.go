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
 *
 * @microservice: core-data-go library
 * @author: Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package data

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go/core/data/clients"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/gorilla/mux"
)

// Put event on the message queue to be processed by the rules engine
func putEventOnQueue(e models.Event) {
	loggingClient.Info("Putting event on message queue", "")
	//	Have multiple implementations (start with ZeroMQ)
	err := ep.SendEventMessage(e)
	if err != nil {
		loggingClient.Error("Unable to send message for event: " + e.String())
	}
}

// Update when the device was last reported connected
func updateDeviceLastReportedConnected(device string) {
	// Config set to skip update last reported
	if !configuration.DeviceUpdateLastConnected {
		loggingClient.Debug("Skipping update of device connected/reported times for:  " + device)
		return
	}

	t := time.Now().UnixNano() / int64(time.Millisecond)

	// Get the device by name
	d, err := mdc.DeviceForName(device)
	if err != nil {
		loggingClient.Error("Error getting device " + device + ": " + err.Error())
		return
	}

	// Couldn't find by name
	if &d == nil {
		// Get the device by ID
		d, err = mdc.Device(device)
		if err != nil {
			loggingClient.Error("Error getting device " + device + ": " + err.Error())
			return
		}

		// Couldn't find device
		if &d == nil {
			loggingClient.Error("Error updating device connected/reported times.  Unknown device with identifier of:  " + device)
			return
		}

		// Got device by ID, now update lastReported/Connected by ID
		err = mdc.UpdateLastConnected(d.Id.Hex(), t)
		if err != nil {
			loggingClient.Error("Problems updating last connected value for device: " + d.Id.Hex())
			return
		}
		err = mdc.UpdateLastReported(d.Id.Hex(), t)
		if err != nil {
			loggingClient.Error("Problems updating last reported value for device: " + d.Id.Hex())
		}
		return
	}

	// Found by name, now update lastReported
	err = mdc.UpdateLastConnectedByName(d.Name, t)
	if err != nil {
		loggingClient.Error("Problems updating last connected value for device: " + d.Name)
		return
	}
	err = mdc.UpdateLastReportedByName(d.Name, t)
	if err != nil {
		loggingClient.Error("Problems updating last reported value for device: " + d.Name)
	}
	return
}

// Update when the device service was last reported connected
func updateDeviceServiceLastReportedConnected(device string) {
	if !configuration.ServiceUpdateLastConnected {
		loggingClient.Debug("Skipping update of device service connected/reported times for:  " + device)
		return
	}

	t := time.Now().UnixNano() / int64(time.Millisecond)

	// Get the device
	d, err := mdc.DeviceForName(device)
	if err != nil {
		loggingClient.Error("Error getting device " + device + ": " + err.Error())
		return
	}

	// Couldn't find by name
	if &d == nil {
		d, err = mdc.Device(device)
		if err != nil {
			loggingClient.Error("Error getting device " + device + ": " + err.Error())
			return
		}
		// Couldn't find device
		if &d == nil {
			loggingClient.Error("Error updating device connected/reported times.  Unknown device with identifier of:  " + device)
			return
		}
	}

	// Get the device service
	s := d.Service
	if &s == nil {
		loggingClient.Error("Error updating device service connected/reported times.  Unknown device service in device:  " + d.Id.Hex())
		return
	}

	msc.UpdateLastConnected(s.Service.Id.Hex(), t)
	msc.UpdateLastReported(s.Service.Id.Hex(), t)
}

// Delete the event and readings
func deleteEvent(e models.Event) error {
	for _, reading := range e.Readings {
		if err := dbc.DeleteReadingById(reading.Id.Hex()); err != nil {
			return err
		}
	}
	if err := dbc.DeleteEventById(e.ID.Hex()); err != nil {
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
		loggingClient.Info("Deleting all events from database")

		err := dbc.ScrubAllEvents()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error("Error scrubbing all events/readings: " + err.Error())
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
		events, err := dbc.Events()
		if err != nil {
			loggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		// Check max limit
		if len(events) > configuration.ReadMaxLimit {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			loggingClient.Error(maxExceededString)
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
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error("Error decoding event: " + err.Error())
			return
		}

		loggingClient.Info("Posting Event: " + e.String())

		// Get device from metadata
		deviceFound := true
		// Try by ID
		d, err := mdc.Device(e.Device)
		if err != nil {
			// Try by name
			d, err = mdc.DeviceForName(e.Device)
			if err != nil {
				deviceFound = false
			}
		}
		// Make sure the identifier is the device name
		if deviceFound {
			e.Device = d.Name
		}

		// See if metadata checking is enabled
		if configuration.MetaDataCheck && !deviceFound {
			loggingClient.Error("Device not found for event: "+err.Error(), "")
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if configuration.ValidateCheck {
			loggingClient.Debug("Validation enabled, parsing events")
			for reading := range e.Readings {
				valid, err := isValidValueDescriptor(e.Readings[reading], e)
				if !valid {
					loggingClient.Error("Validation failed: %s", err.Error())
					return
				}
			}
		}

		// Add the readings to the database
		if configuration.PersistData {
			for i, reading := range e.Readings {
				// Check value descriptor
				_, err := dbc.ValueDescriptorByName(reading.Name)
				if err != nil {
					if err == clients.ErrNotFound {
						http.Error(w, "Value descriptor for a reading not found", http.StatusNotFound)
					} else {
						http.Error(w, err.Error(), http.StatusServiceUnavailable)
					}
					loggingClient.Error(err.Error())
					return
				}

				reading.Device = e.Device // Update the device for the reading

				// Add the reading
				id, err := dbc.AddReading(reading)
				if err != nil {
					http.Error(w, err.Error(), http.StatusServiceUnavailable)
					loggingClient.Error(err.Error())
					return
				}

				e.Readings[i].Id = id // Set the ID for referencing later
			}

			// Add the event to the database
			id, err := dbc.AddEvent(&e)
			if err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				loggingClient.Error(err.Error())
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
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error("Error decoding the event: " + err.Error())
			return
		}

		// Check if the event exists
		to, err := dbc.EventById(from.ID.Hex())
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Event not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		loggingClient.Info("Updating event: " + from.ID.Hex())

		// Update the fields
		if from.Device != "" {
			deviceFound := true
			d, err := mdc.Device(from.Device)
			if err != nil {
				d, err = mdc.DeviceForName(from.Device)
				if err != nil {
					deviceFound = false
				}
			}

			// See if we need to check metadata
			if configuration.MetaDataCheck && !deviceFound {
				http.Error(w, "Error updating event: Device "+from.Device+" doesn't exist", http.StatusNotFound)
				loggingClient.Error("Error updating device, device " + from.Device + " doesn't exist")
				return
			}

			if deviceFound {
				to.Device = d.Name
			} else {
				to.Device = from.Device
			}
		}
		if from.Pushed != 0 {
			to.Pushed = from.Pushed
		}
		if from.Origin != 0 {
			to.Origin = from.Origin
		}

		// Update
		if err = dbc.UpdateEvent(to); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
		//encode(true, w)
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
		e, err := dbc.EventById(id)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Event not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
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
		count, err := dbc.EventCount()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error(), "")
			return
		}

		// Return result
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(strconv.Itoa(count)))
		if err != nil {
			loggingClient.Error(err.Error(), "")
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
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Problem unescaping URL: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Get the device
		// Try by ID
		d, err := mdc.Device(id)
		if err != nil {
			// Try by Name
			d, err = mdc.DeviceForName(id)
			if err != nil {
				loggingClient.Error("Device not found for event: "+err.Error(), "")
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
		}

		count, err := dbc.EventCountByDeviceId(d.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
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
		e, err := dbc.EventById(id)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Event not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		loggingClient.Info("Updating event: " + e.ID.Hex())

		e.Pushed = time.Now().UnixNano() / int64(time.Millisecond)
		err = dbc.UpdateEvent(e)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
		//encode(true, w)
		break
	// Delete the event and all of it's readings
	case http.MethodDelete:
		// Check if the event exists
		e, err := dbc.EventById(id)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Event not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		loggingClient.Info("Deleting event: " + e.ID.Hex())

		if err = deleteEvent(e); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
		//encode(true, w)
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
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error unescaping URL: " + err.Error())
		return
	}

	// Convert limit to int
	limitNum, err := strconv.Atoi(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting to integer: " + err.Error())
		return
	}

	// Get the device
	deviceFound := true
	// Try by ID
	d, err := mdc.Device(deviceId)
	if err != nil {
		// Try by Name
		d, err = mdc.DeviceForName(deviceId)
		if err != nil {
			deviceFound = false
		}
	}

	if deviceFound {
		deviceId = d.Name
	}

	// See if you need to check metadata for the device
	if configuration.MetaDataCheck && !deviceFound {
		http.Error(w, "Error getting events for a device: The device '"+deviceId+"' doesn't exist", http.StatusNotFound)
		loggingClient.Error("Error getting readings for a device: The device doesn't exist")
		return
	}

	switch r.Method {
	case http.MethodGet:
		if limitNum > configuration.ReadMaxLimit {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			loggingClient.Error(maxExceededString)
			return
		}

		eventList, err := dbc.EventsForDeviceLimit(deviceId, limitNum)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
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
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error unescaping the URL: " + err.Error())
		return
	}

	// Get the device
	deviceFound := true
	d, err := mdc.Device(deviceId)
	if err != nil {
		d, err = mdc.DeviceForName(deviceId)
		if err != nil {
			deviceFound = false
		}
	}

	if deviceFound {
		deviceId = d.Name
	}

	// See if you need to check metadata
	if configuration.MetaDataCheck && !deviceFound {
		loggingClient.Error("Device not found for event: "+err.Error(), "")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		// Get the events by the device name
		events, err := dbc.EventsForDevice(deviceId)
		if err != nil {
			loggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		loggingClient.Info("Deleting the events for device: " + deviceId)

		// Delete the events
		count := len(events)
		for _, event := range events {
			if err = deleteEvent(event); err != nil {
				loggingClient.Error(err.Error())
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
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
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Problem converting start time: " + err.Error())
		return
	}

	end, err := strconv.ParseInt(vars["end"], 10, 64)
	// Problems converting end time
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Problem converting end time: " + err.Error())
		return
	}

	limit, err := strconv.Atoi(vars["limit"])
	// Problems converting limit
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Problem converting limit: " + strconv.Itoa(limit))
		return
	}

	switch r.Method {
	case http.MethodGet:
		if limit > configuration.ReadMaxLimit {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			loggingClient.Error(maxExceededString)
			return
		}

		e, err := dbc.EventsByCreationTime(start, end, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
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
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Problem unescaping value descriptor: " + err.Error())
		return
	}

	deviceId, err := url.QueryUnescape(vars["deviceId"])
	// Problems unescaping URL
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Problem unescaping device ID: " + err.Error())
		return
	}

	limitNum, err := strconv.Atoi(limit)
	// Problem converting the limit
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Problem converting limit to integer: " + err.Error())
		return
	}
	switch r.Method {
	case http.MethodGet:
		if limitNum > configuration.ReadMaxLimit {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			loggingClient.Error(maxExceededString)
			return
		}

		// Get the device
		deviceFound := true
		// Try by id
		d, err := mdc.Device(deviceId)
		if err != nil {
			// Try by name
			d, err = mdc.DeviceForName(deviceId)
			if err != nil {
				deviceFound = false
			}
		}

		if deviceFound {
			deviceId = d.Name
		}

		// See if you need to check metadata
		if configuration.MetaDataCheck && !deviceFound {
			loggingClient.Error("Device not found for event: "+err.Error(), "")
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// Get all the events for the device
		e, err := dbc.EventsForDevice(deviceId)
		if err != nil {
			loggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
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
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the age to an integer")
		return
	}

	switch r.Method {
	case http.MethodDelete:
		// Get the events
		events, err := dbc.EventsOlderThanAge(age)
		if err != nil {
			loggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		// Delete all the events
		count := len(events)
		for _, event := range events {
			if err = deleteEvent(event); err != nil {
				loggingClient.Error(err.Error())
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		}

		loggingClient.Info("Deleting events by age: " + vars["age"])

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
		loggingClient.Info("Scrubbing events.  Deleting all events that have been pushed")

		// Get the events
		events, err := dbc.EventsPushed()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}

		// Delete all the events
		count := len(events)
		for _, event := range events {
			if err = deleteEvent(event); err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				loggingClient.Error(err.Error())
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.Itoa(count)))
	}
}
