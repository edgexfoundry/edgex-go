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
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/edgexfoundry/consul-client-go"
	"github.com/edgexfoundry/core-clients-go/metadataclients"
	"github.com/edgexfoundry/core-data-go/clients"
	"github.com/edgexfoundry/core-data-go/messaging"
	"github.com/edgexfoundry/core-domain-go/models"
	"github.com/edgexfoundry/support-logging-client-go"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	formatSpecifier          = "%(\\d+\\$)?([-#+ 0,(\\<]*)?(\\d+)?(\\.\\d+)?([tT])?([a-zA-Z%])"
	maxExceededString string = "Error, exceeded the max limit as defined in config"
)

// Global variables
var dbc *clients.DBClient
var loggingClient logger.LoggingClient
var ep *messaging.EventPublisher
var mdc metadataclients.DeviceClient
var msc metadataclients.ServiceClient

// Helper function for encoding things for returning from REST calls
func encode(i interface{}, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	err := enc.Encode(i)
	// Problems encoding
	if err != nil {
		loggingClient.Error("Error encoding the data: " + err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
}

// Check metadata if the device exists
func checkDevice(device string) bool {
	// First check by name
	_, err := mdc.DeviceForName(device)
	if err != nil {
		// Then check by ID
		_, err = mdc.Device(device)
		if err != nil {
			loggingClient.Error("Can't find device: " + device)
			return false
		}
		return true
	}

	return true
}

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
	if !configuration.Deviceupdatelastconnected {
		loggingClient.Debug("Skipping update of device connected/reported times for:  " + device)
		return
	}

	t := time.Now().Unix()

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
	if !configuration.Serviceupdatelastconnected {
		loggingClient.Debug("Skipping update of device service connected/reported times for:  " + device)
		return
	}

	t := time.Now().Unix()

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

// Printing function purely for debugging purposes
// Print the body of a request to the console
func printBody(r io.ReadCloser) {
	body, err := ioutil.ReadAll(r)
	bodyString := string(body)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(bodyString)
}

/*
Handler for the event API
Status code 404 - event not found
Status code 413 - number of events exceeds limit
Status code 503 - unanticipated issues
api/v1/event
*/
func eventHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	// Get all events
	case "GET":
		events, err := dbc.Events()
		if err != nil {
			loggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		// Check max limit
		if len(events) > configuration.Readmaxlimit {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			loggingClient.Error(maxExceededString)
			return
		}

		encode(events, w)
		break
	// Post a new event
	case "POST":
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
		if configuration.Metadatacheck && !deviceFound {
			loggingClient.Error("Device not found for event: "+err.Error(), "")
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// Add the readings to the database
		if configuration.Persistdata {
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
	case "PUT":
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
			if configuration.Metadatacheck && !deviceFound {
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
		encode(true, w)
	}
}

//GET
//Return the event specified by the event ID
///api/v1/event/{id}
//id - ID of the event to return
func getEventByIdHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case "GET":
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
	case "GET":
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
	case "GET":
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
	case "PUT":
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

		e.Pushed = time.Now().Unix()
		err = dbc.UpdateEvent(e)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}
		break
	// Delete the event and all of it's readings
	case "DELETE":
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

		encode(true, w)
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
	if configuration.Metadatacheck && !deviceFound {
		http.Error(w, "Error getting events for a device: The device '"+deviceId+"' doesn't exist", http.StatusNotFound)
		loggingClient.Error("Error getting readings for a device: The device doesn't exist")
		return
	}

	switch r.Method {
	case "GET":
		if limitNum > configuration.Readmaxlimit {
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
	if configuration.Metadatacheck && !deviceFound {
		loggingClient.Error("Device not found for event: "+err.Error(), "")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch r.Method {
	case "DELETE":
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
	case "GET":
		if limit > configuration.Readmaxlimit {
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
	case "GET":
		if limitNum > configuration.Readmaxlimit {
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
		if configuration.Metadatacheck && !deviceFound {
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
	case "DELETE":
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
	case "DELETE":
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

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.Itoa(count)))
	}
}

// Undocumented feature to remove all readings and events from the database
// This should primarily be used for debugging purposes
func scrubAllHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case "DELETE":
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

// Reading handler
// GET, PUT, and POST readings
func readingHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case "GET":
		r, err := dbc.Readings()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}

		// Check max limit
		if len(r) > configuration.Readmaxlimit {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			loggingClient.Error(maxExceededString)
			return
		}

		encode(r, w)
	case "POST":
		reading := models.Reading{}
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&reading)

		// Problem decoding
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error("Error decoding the reading: " + err.Error())
			return
		}

		// Check the value descriptor
		_, err = dbc.ValueDescriptorByName(reading.Name)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Value descriptor not found for reading", http.StatusConflict)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		// Check device
		if reading.Device != "" {
			// Try by name
			d, err := mdc.DeviceForName(reading.Device)
			// Try by ID
			if err != nil {
				d, err = mdc.Device(reading.Device)
				if err != nil {
					err = errors.New("Device not found for reading")
					loggingClient.Error(err.Error(), "")
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
			}
			reading.Device = d.Name
		}

		if configuration.Persistdata {
			id, err := dbc.AddReading(reading)
			if err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				loggingClient.Error(err.Error())
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(id.Hex()))
		} else {
			// Didn't save the reading in the database
			encode("unsaved", w)
		}
	case "PUT":
		from := models.Reading{}
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&from)

		// Problem decoding
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error("Error decoding the reading: " + err.Error())
			return
		}

		// Check if the reading exists
		to, err := dbc.ReadingById(from.Id.Hex())
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Reading not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		//Update the fields
		if from.Value != "" {
			to.Value = from.Value
		}
		if from.Name != "" {
			_, err := dbc.ValueDescriptorByName(from.Name)
			if err != nil {
				if err == clients.ErrNotFound {
					http.Error(w, "Value descriptor not found for reading", http.StatusConflict)
				} else {
					http.Error(w, err.Error(), http.StatusServiceUnavailable)
				}
				loggingClient.Error(err.Error())
				return
			}
			to.Name = from.Name
		}
		if from.Origin != 0 {
			to.Origin = from.Origin
		}

		err = dbc.UpdateReading(to)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}

		encode(true, w)
	}
}

// Get a reading by id
// HTTP 404 not found if the reading can't be found by the ID
// api/v1/reading/{id}
func getReadingByIdHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	id := vars["id"]

	switch r.Method {
	case "GET":
		reading, err := dbc.ReadingById(id)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Reading not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(reading, w)
	}
}

// Return a count for the number of readings in core data
// api/v1/reading/count
func readingCountHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case "GET":
		count, err := dbc.ReadingCount()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(strconv.Itoa(count)))
		if err != nil {
			loggingClient.Error(err.Error(), "")
		}
	}
}

// Delete a reading by its id
// api/v1/reading/id/{id}
func deleteReadingByIdHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	id := vars["id"]

	switch r.Method {
	case "DELETE":
		// Check if the reading exists
		reading, err := dbc.ReadingById(id)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Reading not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		err = dbc.DeleteReadingById(reading.Id.Hex())
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}

		encode(true, w)
	}
}

// Get all the readings for the device - sort by creation date
// 404 - device ID or name doesn't match
// 413 - max count exceeded
// api/v1/reading/device/{deviceId}/{limit}
func readingByDeviceHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	limit, err := strconv.Atoi(vars["limit"])
	// Problems converting limit to int
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the limit to an integer: " + err.Error())
		return
	}
	deviceId, err := url.QueryUnescape(vars["deviceId"])
	// Problems unescaping URL
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error unescaping the device ID: " + err.Error())
		return
	}

	switch r.Method {
	case "GET":
		if limit > configuration.Readmaxlimit {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			loggingClient.Error(maxExceededString)
			return
		}

		// Try to get device
		// First check by name
		var d models.Device
		d, err := mdc.DeviceForName(deviceId)
		if err != nil {
			// Then check by ID
			d, err = mdc.Device(deviceId)
			if err != nil {
				if configuration.Metadatacheck {
					http.Error(w, "Device doesn't exist for the reading", http.StatusNotFound)
					loggingClient.Error("Error getting readings for a device: The device doesn't exist")
					return
				}
			}
		}

		readings, err := dbc.ReadingsByDevice(d.Name, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}

		encode(readings, w)
	}
}

// Return a list of readings associated with a value descriptor, limited by limit
// HTTP 413 (limit exceeded) if the limit is greater than max limit
// api/v1/reading/name/{name}/{limit}
func readingbyValueDescriptorHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	name, err := url.QueryUnescape(vars["name"])
	// Problems with unescaping URL
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error unescaping value descriptor name: " + err.Error())
		return
	}
	limit, err := strconv.Atoi(vars["limit"])
	// Problems converting limit to int
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the limit to an integer: " + err.Error())
		return
	}

	// Check for value descriptor
	_, err = dbc.ValueDescriptorByName(name)
	if err != nil {
		if err == clients.ErrNotFound {
			loggingClient.Error("Value Descriptor not found for Reading", "")
			http.Error(w, "Value Descriptor not found for Reading", http.StatusNotFound)
			return
		}
	}

	// Limit is too large
	if limit > configuration.Readmaxlimit {
		http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
		loggingClient.Error(maxExceededString)
		return
	}

	read, err := dbc.ReadingsByValueDescriptor(name, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error())
		return
	}

	encode(read, w)
}

// Return a list of readings based on the UOM label for the value decriptor
// api/v1/reading/uomlabel/{uomLabel}/{limit}
func readingByUomLabelHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)

	uomLabel, err := url.QueryUnescape(vars["uomLabel"])
	// Problems unescaping URL
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error unescaping the UOM Label: " + err.Error())
		return
	}

	limit, err := strconv.Atoi(vars["limit"])
	// Problems converting limit to int
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the limit to an integer: " + err.Error())
		return
	}

	// Limit was exceeded
	if limit > configuration.Readmaxlimit {
		http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
		loggingClient.Error(maxExceededString)
		return
	}

	// Get the value descriptors
	vList, err := dbc.ValueDescriptorsByUomLabel(uomLabel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error())
		return
	}

	var vNames []string
	for _, v := range vList {
		vNames = append(vNames, v.Name)
	}

	readings, err := dbc.ReadingsByValueDescriptorNames(vNames, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error())
		return
	}

	encode(readings, w)
}

// Get readings by the value descriptor (specified by the label)
// 413 - limit exceeded
// api/v1/reading/label/{label}/{limit}
func readingByLabelHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	label, err := url.QueryUnescape(vars["label"])
	// Problem unescaping
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error unescaping the label of the value descriptor: " + err.Error())
		return
	}
	limit, err := strconv.Atoi(vars["limit"])
	// Problems converting to int
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the limit to an integer: " + err.Error())
		return
	}

	// Limit is too large
	if limit > configuration.Readmaxlimit {
		loggingClient.Error(maxExceededString)
		http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
		return
	}

	// Get the value descriptors
	vdList, err := dbc.ValueDescriptorsByLabel(label)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error())
		return
	}
	var vdNames []string
	for _, vd := range vdList {
		vdNames = append(vdNames, vd.Name)
	}

	readings, err := dbc.ReadingsByValueDescriptorNames(vdNames, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error())
		return
	}

	encode(readings, w)
}

// Return a list of readings who's value descriptor has the type
// 413 - number exceeds the current limit
// /reading/type/{type}/{limit}
func readingByTypeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)

	t, err := url.QueryUnescape(vars["type"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error escaping the type: " + err.Error())
		return
	}

	l, err := strconv.Atoi(vars["limit"])
	// Problem converting to int
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the limit to an integer: " + err.Error())
		return
	}

	// Limit exceeds max limit
	if l > configuration.Readmaxlimit {
		http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
		loggingClient.Error(maxExceededString)
		return
	}

	// Get the value descriptors
	vdList, err := dbc.ValueDescriptorsByType(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error())
		return
	}
	var vdNames []string
	for _, vd := range vdList {
		vdNames = append(vdNames, vd.Name)
	}

	readings, err := dbc.ReadingsByValueDescriptorNames(vdNames, l)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error())
		return
	}

	encode(readings, w)
}

// Return a list of readings between the start and end (creation time)
// /reading/{start}/{end}/{limit}
func readingByCreationTimeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	s, err := strconv.ParseInt((vars["start"]), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the start time to an integer: " + err.Error())
		return
	}
	e, err := strconv.ParseInt((vars["end"]), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the end time to an integer: " + err.Error())
		return
	}
	l, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the limit to an integer: " + err.Error())
		return
	}

	switch r.Method {
	case "GET":
		if l > configuration.Readmaxlimit {
			loggingClient.Error(maxExceededString)
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			return
		}

		readings, err := dbc.ReadingsByCreationTime(s, e, l)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}

		encode(readings, w)
	}
}

// Return a list of redings associated with the device and value descriptor
// Limit exceeded exception 413 if the limit exceeds the max limit
// api/v1/reading/name/{name}/device/{device}/{limit}
func readingByValueDescriptorAndDeviceHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)

	// Get the variables from the URL
	name, err := url.QueryUnescape(vars["name"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error unescaping the value descriptor name: " + err.Error())
		return
	}

	device, err := url.QueryUnescape(vars["device"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error unescaping the device: " + err.Error())
		return
	}

	limit, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting limit to an integer: " + err.Error())
		return
	}

	if limit > configuration.Readmaxlimit {
		loggingClient.Error(maxExceededString)
		http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
		return
	}

	// Check for the device
	if !checkDevice(device) {
		loggingClient.Error("Device not found for reading", "")
		http.Error(w, "Device not found for reading", http.StatusNotFound)
		return
	}

	// Check for value descriptor
	_, err = dbc.ValueDescriptorByName(name)
	if err != nil {
		if err == clients.ErrNotFound {
			loggingClient.Error("Value Descriptor not found for Reading", "")
			http.Error(w, "Value Descriptor not found for Reading", http.StatusNotFound)
			return
		}
	}

	readings, err := dbc.ReadingsByDeviceAndValueDescriptor(device, name, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error())
		return
	}

	encode(readings, w)
}

// Check if the value descriptor matches the format string regular expression
func validateFormatString(v models.ValueDescriptor) (bool, error) {
	// No formatting specified
	if v.Formatting == "" {
		return true, nil
	} else {
		return regexp.MatchString(formatSpecifier, v.Formatting)
	}
}

// GET, POST, and PUT for value descriptors
// api/v1/valuedescriptor
func valueDescriptorHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case "GET":
		vList, err := dbc.ValueDescriptors()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}

		// Check the limit
		if len(vList) > configuration.Readmaxlimit {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			loggingClient.Error(maxExceededString)
			return
		}

		encode(vList, w)
	case "POST":
		dec := json.NewDecoder(r.Body)
		v := models.ValueDescriptor{}
		err := dec.Decode(&v)
		// Problems decoding
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error("Error decoding the value descriptor: " + err.Error())
			return
		}

		// Check the formatting
		match, err := validateFormatString(v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error("Error checking for format string for POSTed value descriptor")
			return
		}
		if !match {
			err := errors.New("Error posting value descriptor. Format is not a valid printf format.")
			http.Error(w, err.Error(), http.StatusConflict)
			loggingClient.Error(err.Error())
			return
		}

		id, err := dbc.AddValueDescriptor(v)
		if err != nil {
			if err == clients.ErrNotUnique {
				http.Error(w, "Value Descriptor already exists", http.StatusConflict)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(id.Hex()))
	case "PUT":
		dec := json.NewDecoder(r.Body)
		from := models.ValueDescriptor{}
		err := dec.Decode(&from)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error("Error decoding the value descriptor: " + err.Error())
			return
		}

		// Find the value descriptor thats being updated
		// Try by ID
		to, err := dbc.ValueDescriptorById(from.Id.Hex())
		if err != nil {
			to, err = dbc.ValueDescriptorByName(from.Name)
			if err != nil {
				if err == clients.ErrNotFound {
					http.Error(w, "Value descriptor not found", http.StatusNotFound)
				} else {
					http.Error(w, err.Error(), http.StatusServiceUnavailable)
				}
				loggingClient.Error(err.Error())
				return
			}
		}

		// Update the fields
		if from.DefaultValue != "" {
			to.DefaultValue = from.DefaultValue
		}
		if from.Formatting != "" {
			match, err := regexp.MatchString(formatSpecifier, from.Formatting)
			if err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				loggingClient.Error("Error checking formatting for updated value descriptor")
				return
			}
			if !match {
				http.Error(w, "Value descriptor's format string doesn't fit the required pattern", http.StatusConflict)
				loggingClient.Error("Value descriptor's format string doesn't fit the required pattern: " + formatSpecifier)
				return
			}
			to.Formatting = from.Formatting
		}
		if from.Labels != nil {
			to.Labels = from.Labels
		}

		if from.Max != "" {
			to.Max = from.Max
		}
		if from.Min != "" {
			to.Min = from.Min
		}
		if from.Name != "" {
			// Check if value descriptor is still in use by readings if the name changes
			if from.Name != to.Name {
				r, err := dbc.ReadingsByValueDescriptor(to.Name, 10) // Arbitrary limit, we're just checking if there are any readings
				if err != nil {
					http.Error(w, err.Error(), http.StatusServiceUnavailable)
					loggingClient.Error("Error checking the readings for the value descriptor: " + err.Error())
					return
				}
				// Value descriptor is still in use
				if len(r) != 0 {
					http.Error(w, "Data integrity issue. Value Descriptor still in use by readings", http.StatusConflict)
					loggingClient.Error("Data integrity issue.  Value Descriptor with name:  " + from.Name + " is still referenced by existing readings.")
					return
				}
			}
			to.Name = from.Name
		}
		if from.Origin != 0 {
			to.Origin = from.Origin
		}
		if from.Type != "" {
			to.Type = from.Type
		}
		if from.UomLabel != "" {
			to.UomLabel = from.UomLabel
		}

		// Push the updated valuedescriptor to the database
		err = dbc.UpdateValueDescriptor(to)
		if err != nil {
			if err == clients.ErrNotUnique {
				http.Error(w, "Value descriptor name is not unique", http.StatusConflict)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(true, w)
	}
}

// Delete the value descriptor based on the ID
// DataValidationException (HTTP 409) - The value descriptor is still referenced by readings
// NotFoundException (404) - Can't find the value descriptor
// valuedescriptor/id/{id}
func deleteValueDescriptorByIdHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	id := vars["id"]

	// Check if the value descriptor exists
	vd, err := dbc.ValueDescriptorById(id)
	if err != nil {
		if err == clients.ErrNotFound {
			http.Error(w, "Value descriptor not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error())
		return
	}

	if err = deleteValueDescriptor(vd, w); err != nil {
		return
	}

	encode(true, w)
}

// Value descriptors based on name
// api/v1/valuedescriptor/name/{name}
func valueDescriptorByNameHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	name, err := url.QueryUnescape(vars["name"])

	// Problems unescaping
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error unescaping the value descriptor name: " + err.Error())
		return
	}

	switch r.Method {
	case "GET":
		v, err := dbc.ValueDescriptorByName(name)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Value Descriptor not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(v, w)
	case "DELETE":
		// Check if the value descriptor exists
		vd, err := dbc.ValueDescriptorByName(name)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Value Descriptor not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		if err = deleteValueDescriptor(vd, w); err != nil {
			return
		}

		encode(true, w)
	}
}

func deleteValueDescriptor(vd models.ValueDescriptor, w http.ResponseWriter) error {
	// Check if the value descriptor is still in use by readings
	readings, err := dbc.ReadingsByValueDescriptor(vd.Name, 10)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error())
		return err
	}
	if len(readings) > 0 {
		err = errors.New("Data integrity issue.  Value Descriptor is still referenced by existing readings.")
		http.Error(w, err.Error(), http.StatusConflict)
		loggingClient.Error(err.Error())
		return err
	}

	// Delete the value descriptor
	if err = dbc.DeleteValueDescriptorById(vd.Id.Hex()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error())
		return err
	}

	return nil
}

// Get a value descriptor based on the ID
// HTTP 404 not found if the ID isn't in the database
// api/v1/valuedescriptor/{id}
func valueDescriptorByIdHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	id := vars["id"]

	switch r.Method {
	case "GET":
		v, err := dbc.ValueDescriptorById(id)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Value descriptor not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(v, w)
	}
}

// Get the value descriptor from the UOM label
// api/v1/valuedescriptor/uomlabel/{uomLabel}
func valueDescriptorByUomLabelHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	uomLabel, err := url.QueryUnescape(vars["uomLabel"])

	// Prolem unescaping
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error unescaping the UOM Label of the value descriptor: " + err.Error())
		return
	}

	switch r.Method {
	case "GET":
		v, err := dbc.ValueDescriptorsByUomLabel(uomLabel)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}

		encode(v, w)
	}
}

// Get value descriptors who have one of the labels
// api/v1/valuedescriptor/label/{label}
func valueDescriptorByLabelHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	label, err := url.QueryUnescape(vars["label"])

	// Problem unescaping
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error unescaping label for the value descriptor: " + err.Error())
		return
	}

	switch r.Method {
	case "GET":
		v, err := dbc.ValueDescriptorsByLabel(label)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}

		encode(v, w)
	}
}

// Return the value descriptors that are asociated with a device
// The value descriptor is expected parameters on puts or expected values on get/put commands
// api/v1/valuedescriptor/devicename/{device}
func valueDescriptorByDeviceHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)

	device, err := url.QueryUnescape(vars["device"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error unescaping the device: " + err.Error())
		return
	}

	// Get the device
	d, err := mdc.DeviceForName(device)
	if err != nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		loggingClient.Error("Device not found: " + err.Error())
		return
	}

	// Get the value descriptors
	vdList, err := valueDescriptorsForDevice(d, w)
	if err != nil {
		return
	}

	encode(vdList, w)
}

// Return the value descriptors that are associated with the device specified by the device ID
// Associated value descripts are expected parameters of PUT commands and expected results of PUT/GET commands
// api/v1/valuedescriptor/deviceid/{id}
func valueDescriptorByDeviceIdHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)

	deviceId, err := url.QueryUnescape(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error unescaping the device ID: " + err.Error())
		return
	}

	// Get the device
	d, err := mdc.Device(deviceId)
	if err != nil {
		if err == metadataclients.ErrNotFound {
			http.Error(w, "Device not found: "+err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Problem getting device from metadata: "+err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error("Device not found: " + err.Error())
		return
	}

	// Get the value descriptors
	vdList, err := valueDescriptorsForDevice(d, w)
	if err != nil {
		return
	}

	encode(vdList, w)
}

// Get the value descriptors for the device
func valueDescriptorsForDevice(d models.Device, w http.ResponseWriter) ([]models.ValueDescriptor, error) {
	// Get the names of the value descriptors
	vdNames := []string{}
	d.AllAssociatedValueDescriptors(&vdNames)

	// Get the value descriptors
	vdList := []models.ValueDescriptor{}
	for _, name := range vdNames {
		vd, err := dbc.ValueDescriptorByName(name)

		// Not an error if not found
		if err == clients.ErrNotFound {
			continue
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return vdList, err
		}

		vdList = append(vdList, vd)
	}

	return vdList, nil
}

// Test if the service is working
func pingHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	_, err := w.Write([]byte("pong"))
	if err != nil {
		loggingClient.Error("Error writing pong: " + err.Error())
	}
}

func loadRestRoutes() *mux.Router {
	r := mux.NewRouter()
	baseUrl := "/api/v1"

	// EVENTS
	r.HandleFunc(baseUrl+"/event", eventHandler).Methods("GET", "PUT", "POST")
	r.HandleFunc(baseUrl+"/event/scrub", scrubHandler).Methods("DELETE")
	r.HandleFunc(baseUrl+"/event/scruball", scrubAllHandler).Methods("DELETE")
	r.HandleFunc(baseUrl+"/event/count", eventCountHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/event/count/{deviceId}", eventCountByDeviceIdHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/event/{id}", getEventByIdHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/event/id/{id}", eventIdHandler).Methods("DELETE", "PUT")
	r.HandleFunc(baseUrl+"/event/device/{deviceId}/{limit:[0-9]+}", getEventByDeviceHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/event/device/{deviceId}", deleteByDeviceIdHandler).Methods("DELETE")
	r.HandleFunc(baseUrl+"/event/removeold/age/{age:[0-9]+}", eventByAgeHandler).Methods("DELETE")
	r.HandleFunc(baseUrl+"/event/{start:[0-9]+}/{end:[0-9]+}/{limit:[0-9]+}", eventByCreationTimeHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/event/device/{deviceId}/valuedescriptor/{valueDescriptor}/{limit:[0-9]+}", readingByDeviceFilteredValueDescriptor).Methods("GET")

	// READINGS
	r.HandleFunc(baseUrl+"/reading", readingHandler).Methods("GET", "PUT", "POST")
	r.HandleFunc(baseUrl+"/reading/count", readingCountHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/reading/id/{id}", deleteReadingByIdHandler).Methods("DELETE")
	r.HandleFunc(baseUrl+"/reading/{id}", getReadingByIdHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/reading/device/{deviceId}/{limit:[0-9]+}", readingByDeviceHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/reading/name/{name}/{limit:[0-9]+}", readingbyValueDescriptorHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/reading/uomlabel/{uomLabel}/{limit:[0-9]+}", readingByUomLabelHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/reading/label/{label}/{limit:[0-9]+}", readingByLabelHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/reading/type/{type}/{limit:[0-9]+}", readingByTypeHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/reading/{start:[0-9]+}/{end:[0-9]+}/{limit:[0-9]+}", readingByCreationTimeHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/reading/name/{name}/device/{device}/{limit:[0-9]+}", readingByValueDescriptorAndDeviceHandler).Methods("GET")

	// VALUE DESCRIPTORS
	r.HandleFunc(baseUrl+"/valuedescriptor", valueDescriptorHandler).Methods("GET", "PUT", "POST")
	r.HandleFunc(baseUrl+"/valuedescriptor/id/{id}", deleteValueDescriptorByIdHandler).Methods("DELETE")
	r.HandleFunc(baseUrl+"/valuedescriptor/name/{name}", valueDescriptorByNameHandler).Methods("GET", "DELETE")
	r.HandleFunc(baseUrl+"/valuedescriptor/{id}", valueDescriptorByIdHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/valuedescriptor/uomlabel/{uomLabel}", valueDescriptorByUomLabelHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/valuedescriptor/label/{label}", valueDescriptorByLabelHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/valuedescriptor/devicename/{device}", valueDescriptorByDeviceHandler).Methods("GET")
	r.HandleFunc(baseUrl+"/valuedescriptor/deviceid/{id}", valueDescriptorByDeviceIdHandler).Methods("GET")

	// Ping Resource
	r.HandleFunc(baseUrl+"/ping", pingHandler)

	return r
}

// Heartbeat for the data microservice - send a message to logging service
func heartbeat() {
	// Loop forever
	for true {
		loggingClient.Info(configuration.Heartbeatmsg, "")
		time.Sleep(time.Millisecond * time.Duration(configuration.Heartbeattime)) // Sleep based on configuration
	}
}

// Read the configuration file and update configuration struct
func readConfigurationFile(path string) error {
	// Read the configuration file
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading configuration file: " + err.Error())
		return err
	}

	// Decode the configuration as JSON
	err = json.Unmarshal(contents, &configuration)
	if err != nil {
		fmt.Println("Error reading configuration file: " + err.Error())
		return err
	}

	return nil
}

func main() {
	start := time.Now()

	// Load configuration data
	readConfigurationFile("./res/configuration.json")

	// Create Logger (Default Parameters)
	loggingClient = logger.NewClient(configuration.Applicationname, configuration.Loggingremoteurl)
	loggingClient.LogFilePath = configuration.Loggingfile

	// Initialize service on Consul
	err := consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    configuration.Applicationname,
		ServicePort:    configuration.Serverport,
		ServiceAddress: configuration.Applicationname,
		CheckAddress:   configuration.Consulcheckaddress,
		CheckInterval:  "10s",
		ConsulAddress:  configuration.Consulhost,
		ConsulPort:     configuration.Consulport,
	})

	if err != nil {
		loggingClient.Error("Connection to Consul could not be made: "+err.Error(), "")
	}

	// Update configuration data from Consul
	if err := consulclient.CheckKeyValuePairs(&configuration, configuration.Applicationname, strings.Split(configuration.Consulprofilesactive, ";")); err != nil {
		loggingClient.Error("Error getting key/values from Consul: "+err.Error(), "")
	}

	// Create a database client
	dbc, err = clients.NewDBClient(clients.DBConfiguration{
		DbType:       clients.MONGO,
		Host:         configuration.Datamongodbhost,
		Port:         configuration.Datamongodbport,
		Timeout:      configuration.DatamongodbsocketTimeout,
		DatabaseName: configuration.Datamongodbdatabase,
		Username:     configuration.Datamongodbusername,
		Password:     configuration.Datamongodbpassword,
	})
	if err != nil {
		loggingClient.Error("Couldn't connect to database: "+err.Error(), "")
		return
	}

	// Create metadata clients
	mdc = metadataclients.NewDeviceClient(configuration.Metadbdeviceurl)
	msc = metadataclients.NewServiceClient(configuration.Metadbdeviceserviceurl)

	// Create the event publisher
	ep = messaging.NewZeroMQPublisher(messaging.ZeroMQConfiguration{
		AddressPort: configuration.Zeromqaddressport,
	})

	// Start heartbeat
	go heartbeat()

	r := loadRestRoutes()
	http.TimeoutHandler(nil, time.Millisecond*time.Duration(5000), "Request timed out")
	loggingClient.Info(configuration.Appopenmsg, "")

	// Time it took to start service
	loggingClient.Info("Service started in: "+time.Since(start).String(), "")
	loggingClient.Info("Listening on port: " + strconv.Itoa(configuration.Serverport))

	loggingClient.Error(http.ListenAndServe(":"+strconv.Itoa(configuration.Serverport), r).Error())
}
