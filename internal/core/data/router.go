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

	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
)

func LoadRestRoutes() *mux.Router {
	r := mux.NewRouter()
	b := r.PathPrefix("/api/v1").Subrouter()

	// EVENTS
	// /api/v1/event
	b.HandleFunc("/event", eventHandler).Methods(http.MethodGet, http.MethodPut, http.MethodPost)
	e := b.PathPrefix("/event").Subrouter()
	e.HandleFunc("/scrub", scrubHandler).Methods(http.MethodDelete)
	e.HandleFunc("/scruball", scrubAllHandler).Methods(http.MethodDelete)
	e.HandleFunc("/count", eventCountHandler).Methods(http.MethodGet)
	e.HandleFunc("/count/{deviceId}", eventCountByDeviceIdHandler).Methods(http.MethodGet)
	e.HandleFunc("/{id}", getEventByIdHandler).Methods(http.MethodGet)
	e.HandleFunc("/id/{id}", eventIdHandler).Methods(http.MethodDelete, http.MethodPut)
	e.HandleFunc("/device/{deviceId}/{limit:[0-9]+}", getEventByDeviceHandler).Methods(http.MethodGet)
	e.HandleFunc("/device/{deviceId}", deleteByDeviceIdHandler).Methods(http.MethodDelete)
	e.HandleFunc("/removeold/age/{age:[0-9]+}", eventByAgeHandler).Methods(http.MethodDelete)
	e.HandleFunc("/{start:[0-9]+}/{end:[0-9]+}/{limit:[0-9]+}", eventByCreationTimeHandler).Methods(http.MethodGet)
	e.HandleFunc("/device/{deviceId}/valuedescriptor/{valueDescriptor}/{limit:[0-9]+}", readingByDeviceFilteredValueDescriptor).Methods(http.MethodGet)

	// READINGS
	// /api/v1/reading
	b.HandleFunc("/reading", readingHandler).Methods(http.MethodGet, http.MethodPut, http.MethodPost)
	rd := b.PathPrefix("/reading").Subrouter()
	rd.HandleFunc("/count", readingCountHandler).Methods(http.MethodGet)
	rd.HandleFunc("/id/{id}", deleteReadingByIdHandler).Methods(http.MethodDelete)
	rd.HandleFunc("/{id}", getReadingByIdHandler).Methods(http.MethodGet)
	rd.HandleFunc("/device/{deviceId}/{limit:[0-9]+}", readingByDeviceHandler).Methods(http.MethodGet)
	rd.HandleFunc("/name/{name}/{limit:[0-9]+}", readingbyValueDescriptorHandler).Methods(http.MethodGet)
	rd.HandleFunc("/uomlabel/{uomLabel}/{limit:[0-9]+}", readingByUomLabelHandler).Methods(http.MethodGet)
	rd.HandleFunc("/label/{label}/{limit:[0-9]+}", readingByLabelHandler).Methods(http.MethodGet)
	rd.HandleFunc("/type/{type}/{limit:[0-9]+}", readingByTypeHandler).Methods(http.MethodGet)
	rd.HandleFunc("/{start:[0-9]+}/{end:[0-9]+}/{limit:[0-9]+}", readingByCreationTimeHandler).Methods(http.MethodGet)
	rd.HandleFunc("/name/{name}/device/{device}/{limit:[0-9]+}", readingByValueDescriptorAndDeviceHandler).Methods(http.MethodGet)

	// VALUE DESCRIPTORS
	// /api/v1/valuedescriptor
	b.HandleFunc("/valuedescriptor", valueDescriptorHandler).Methods(http.MethodGet, http.MethodPut, http.MethodPost)
	vd := b.PathPrefix("/valuedescriptor").Subrouter()
	vd.HandleFunc("/id/{id}", deleteValueDescriptorByIdHandler).Methods(http.MethodDelete)
	vd.HandleFunc("/name/{name}", valueDescriptorByNameHandler).Methods(http.MethodGet, http.MethodDelete)
	vd.HandleFunc("/{id}", valueDescriptorByIdHandler).Methods(http.MethodGet)
	vd.HandleFunc("/uomlabel/{uomLabel}", valueDescriptorByUomLabelHandler).Methods(http.MethodGet)
	vd.HandleFunc("/label/{label}", valueDescriptorByLabelHandler).Methods(http.MethodGet)
	vd.HandleFunc("/devicename/{device}", valueDescriptorByDeviceHandler).Methods(http.MethodGet)
	vd.HandleFunc("/deviceid/{id}", valueDescriptorByDeviceIdHandler).Methods(http.MethodGet)

	// Ping Resource
	// /api/v1/ping
	b.HandleFunc("/ping", pingHandler)

	//Config Resource
	// /api/v1/config
	b.HandleFunc("/config", configHandler)

	return r
}

/*
Return number of events in Core Data
/api/v1/event/count
*/
func eventCountHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	count, err := countEvents()
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

	// Check device
	count, err := countEventsByDevice(id)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("error checking device %s %v", id, err))
		switch err := err.(type) {
		case types.ErrNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		default: //return an error on everything else.
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strconv.Itoa(count)))
}

// Remove all the old events and associated readings (by age)
// event/removeold/age/{age}
func eventByAgeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	age, err := strconv.ParseInt(vars["age"], 10, 64)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error converting the age to an integer")
		return
	}

	LoggingClient.Info("Deleting events by age: " + vars["age"])

	count, err := deleteEventsByAge(age)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strconv.Itoa(count)))
}

/*
Handler for the event API
Status code 404 - event not found
Status code 413 - number of events exceeds limit
Status code 500 - unanticipated issues
api/v1/event
*/
func eventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	switch r.Method {
	// Get all events
	case http.MethodGet:
		events, err := getEvents(Configuration.Service.ReadMaxLimit)
		if err != nil {
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

		newId, err := addNew(e)
		if err != nil {
			switch t := err.(type) {
			case *errors.ErrValueDescriptorNotFound:
				http.Error(w, t.Error(), http.StatusBadRequest)
			case *errors.ErrValueDescriptorInvalid:
				http.Error(w, t.Error(), http.StatusBadRequest)
			default:
				http.Error(w, t.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(newId))

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

		LoggingClient.Info("Updating event: " + from.ID.Hex())
		err = updateEvent(from)
		if err != nil {
			switch t := err.(type) {
			case *errors.ErrEventNotFound:
				http.Error(w, t.Error(), http.StatusNotFound)
			default:
				http.Error(w, t.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
	}
}

// Undocumented feature to remove all readings and events from the database
// This should primarily be used for debugging purposes
func scrubAllHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	LoggingClient.Info("Deleting all events from database")

	err := deleteAllEvents()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
		return
	}

	encode(true, w)
}

//GET
//Return the event specified by the event ID
///api/v1/event/{id}
//id - ID of the event to return
func getEventByIdHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	// URL parameters
	vars := mux.Vars(r)
	id := vars["id"]

	// Get the event
	e, err := getEventById(id)
	if err != nil {
		switch x := err.(type) {
		case *errors.ErrEventNotFound:
			http.Error(w, x.Error(), http.StatusNotFound)
		default:
			http.Error(w, x.Error(), http.StatusInternalServerError)
		}

		LoggingClient.Error(err.Error())
		return
	}

	encode(e, w)
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
	if checkDevice(deviceId) != nil {
		LoggingClient.Error(fmt.Sprintf("error checking device %s %v", deviceId, err))
		switch err := err.(type) {
		case types.ErrNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		default: //return an error on everything else.
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		err := testLimit(limitNum)
		if err != nil {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			return
		}

		eventList, err := getEventsByDeviceIdLimit(limitNum, deviceId)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		encode(eventList, w)
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
	// Set the 'pushed' timestamp for the event to the current time - event is going to another (not EdgeX) service
	case http.MethodPut:
		LoggingClient.Info("Updating event: " + id)

		err := updateEventPushDate(id)
		if err != nil {
			switch x := err.(type) {
			case *errors.ErrEventNotFound:
				http.Error(w, x.Error(), http.StatusNotFound)
			default:
				http.Error(w, x.Error(), http.StatusInternalServerError)
			}

			LoggingClient.Error(err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
		break
		// Delete the event and all of it's readings
	case http.MethodDelete:
		LoggingClient.Info("Deleting event: " + id)
		err := deleteEventById(id)
		if err != nil {
			switch x := err.(type) {
			case *errors.ErrEventNotFound:
				http.Error(w, x.Error(), http.StatusNotFound)
			default:
				http.Error(w, x.Error(), http.StatusInternalServerError)
			}

			LoggingClient.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
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
	if checkDevice(deviceId) != nil {
		LoggingClient.Error(fmt.Sprintf("error checking device %s %v", deviceId, err))
		switch err := err.(type) {
		case types.ErrNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		default: //return an error on everything else.
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
	}

	switch r.Method {
	case http.MethodDelete:
		count, err := deleteEvents(deviceId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
		err := testLimit(limit)
		if err != nil {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			return
		}

		eventList, err := getEventsByCreationTime(limit, start, end)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		encode(eventList, w)
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
		// Check device
		if checkDevice(deviceId) != nil {
			LoggingClient.Error(fmt.Sprintf("error checking device %s %v", deviceId, err))
			switch err := err.(type) {
			case types.ErrNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			default: //return an error on everything else.
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		}

		err := testLimit(limitNum)
		if err != nil {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			return
		}

		readings, err := getReadingsByDeviceId(limitNum, deviceId, valueDescriptor)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
		count, err:= doScrub()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.Itoa(count)))
	}
}
