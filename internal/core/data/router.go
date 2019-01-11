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
	"runtime"
	"strconv"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
)

func LoadRestRoutes() *mux.Router {
	r := mux.NewRouter()

	// Ping Resource
	r.HandleFunc(clients.ApiPingRoute, pingHandler).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(clients.ApiConfigRoute, configHandler).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(clients.ApiMetricsRoute, metricsHandler).Methods(http.MethodGet)

	// Events
	r.HandleFunc(clients.ApiEventRoute, eventHandler).Methods(http.MethodGet, http.MethodPut, http.MethodPost)
	e := r.PathPrefix(clients.ApiEventRoute).Subrouter()
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

	// Readings
	r.HandleFunc(clients.ApiReadingRoute, readingHandler).Methods(http.MethodGet, http.MethodPut, http.MethodPost)
	rd := r.PathPrefix(clients.ApiReadingRoute).Subrouter()
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

	// Value descriptors
	r.HandleFunc(clients.ApiValueDescriptorRoute, valueDescriptorHandler).Methods(http.MethodGet, http.MethodPut, http.MethodPost)
	vd := r.PathPrefix(clients.ApiValueDescriptorRoute).Subrouter()
	vd.HandleFunc("/id/{id}", deleteValueDescriptorByIdHandler).Methods(http.MethodDelete)
	vd.HandleFunc("/name/{name}", valueDescriptorByNameHandler).Methods(http.MethodGet, http.MethodDelete)
	vd.HandleFunc("/{id}", valueDescriptorByIdHandler).Methods(http.MethodGet)
	vd.HandleFunc("/uomlabel/{uomLabel}", valueDescriptorByUomLabelHandler).Methods(http.MethodGet)
	vd.HandleFunc("/label/{label}", valueDescriptorByLabelHandler).Methods(http.MethodGet)
	vd.HandleFunc("/devicename/{device}", valueDescriptorByDeviceHandler).Methods(http.MethodGet)
	vd.HandleFunc("/deviceid/{id}", valueDescriptorByDeviceIdHandler).Methods(http.MethodGet)

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
		LoggingClient.Error(err.Error())
		return
	}

	// Return result
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(strconv.Itoa(count)))
	if err != nil {
		LoggingClient.Error(err.Error())
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
		case *types.ErrServiceClient:
			http.Error(w, err.Error(), err.StatusCode)
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

		newId, err := addNewEvent(e)
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

		LoggingClient.Info("Updating event: " + from.ID)
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
	if err := checkDevice(deviceId); err != nil {
		LoggingClient.Error(fmt.Sprintf("error checking device %s %v", deviceId, err))
		switch err := err.(type) {
		case *types.ErrServiceClient:
			http.Error(w, err.Error(), err.StatusCode)
			return
		default: //return an error on everything else.
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		err := checkMaxLimit(limitNum)
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
	if err := checkDevice(deviceId); err != nil {
		LoggingClient.Error(fmt.Sprintf("error checking device %s %v", deviceId, err))
		switch err := err.(type) {
		case *types.ErrServiceClient:
			http.Error(w, err.Error(), err.StatusCode)
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
		err := checkMaxLimit(limit)
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
		if err := checkDevice(deviceId); err != nil {
			LoggingClient.Error(fmt.Sprintf("error checking device %s %v", deviceId, err))
			switch err := err.(type) {
			case *types.ErrServiceClient:
				http.Error(w, err.Error(), err.StatusCode)
				return
			default: //return an error on everything else.
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		}

		err := checkMaxLimit(limitNum)
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
		count, err := scrubPushedEvents()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.Itoa(count)))
	}
}

// Test if the service is working
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong"))
}

func configHandler(w http.ResponseWriter, _ *http.Request) {
	encode(Configuration, w)
}

// Reading handler
// GET, PUT, and POST readings
func readingHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case http.MethodGet:
		r, err := getAllReadings()

		if err != nil {
			switch err.(type) {
			case *errors.ErrLimitExceeded:
				http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		encode(r, w)
	case http.MethodPost:
		reading, err := decodeReading(r.Body)

		// Problem decoding
		if err != nil {
			switch err.(type) {
			case *errors.ErrJsonDecoding:
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			case *errors.ErrDbNotFound:
				http.Error(w, "Value descriptor not found for reading", http.StatusConflict)
				return
			case *errors.ErrValueDescriptorInvalid:
				http.Error(w, err.Error(), http.StatusConflict)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Check device
		if reading.Device != "" {
			if err := checkDevice(reading.Device); err != nil {
				LoggingClient.Error(fmt.Sprintf("error checking device %s %v", reading.Device, err))
				switch err := err.(type) {
				case *types.ErrServiceClient:
					http.Error(w, err.Error(), err.StatusCode)
					return
				default: //return an error on everything else.
					http.Error(w, err.Error(), http.StatusServiceUnavailable)
					return
				}
			}
		}

		if Configuration.Writable.PersistData {
			id, err := addReading(reading)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(id))
		} else {
			// Didn't save the reading in the database
			encode("unsaved", w)
		}
	case http.MethodPut:
		from, err := decodeReading(r.Body)
		// Problem decoding
		if err != nil {
			switch err.(type) {
			case *errors.ErrJsonDecoding:
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			case *errors.ErrDbNotFound:
				http.Error(w, "Value descriptor not found for reading", http.StatusConflict)
				return
			case *errors.ErrValueDescriptorInvalid:
				http.Error(w, err.Error(), http.StatusConflict)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		err = updateReading(from)
		if err != nil {
			switch err.(type) {
			case *errors.ErrDbNotFound:
				http.Error(w, "Value descriptor not found for reading", http.StatusNotFound)
				return
			case *errors.ErrValueDescriptorInvalid:
				http.Error(w, err.Error(), http.StatusConflict)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
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
	case http.MethodGet:
		reading, err := getReadingById(id)
		if err != nil {
			switch err := err.(type) {
			case *errors.ErrDbNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			default: //return an error on everything else.
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		}

		encode(reading, w)
	}
}

// Return a count for the number of readings in core data
// api/v1/reading/count
func readingCountHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case http.MethodGet:
		count, err := countReadings()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(strconv.Itoa(count)))
		if err != nil {
			LoggingClient.Error(err.Error())
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
	case http.MethodDelete:
		err := deleteReadingById(id)
		if err != nil {
			switch err := err.(type) {
			case *errors.ErrDbNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			default: //return an error on everything else.
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error converting the limit to an integer: " + err.Error())
		return
	}
	deviceId, err := url.QueryUnescape(vars["deviceId"])
	// Problems unescaping URL
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error unescaping the device ID: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		err := checkMaxLimit(limit)
		if err != nil {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			return
		}

		readings, err := getReadingsByDevice(deviceId, limit)
		if err != nil {
			switch err := err.(type) {
			case *types.ErrServiceClient:
				http.Error(w, err.Error(), err.StatusCode)
				return
			default:
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error unescaping value descriptor name: " + err.Error())
		return
	}
	limit, err := strconv.Atoi(vars["limit"])
	// Problems converting limit to int
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error converting the limit to an integer: " + err.Error())
		return
	}

	read, err := getReadingsByValueDescriptor(name, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error unescaping the UOM Label: " + err.Error())
		return
	}

	limit, err := strconv.Atoi(vars["limit"])
	// Problems converting limit to int
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error converting the limit to an integer: " + err.Error())
		return
	}

	// Limit was exceeded
	err = checkMaxLimit(limit)
	if err != nil {
		http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
		return
	}

	// Get the value descriptors
	vList, err := getValueDescriptorsByUomLabel(uomLabel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var vNames []string
	for _, v := range vList {
		vNames = append(vNames, v.Name)
	}

	readings, err := getReadingsByValueDescriptorNames(vNames, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error unescaping the label of the value descriptor: " + err.Error())
		return
	}
	limit, err := strconv.Atoi(vars["limit"])
	// Problems converting to int
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error converting the limit to an integer: " + err.Error())
		return
	}

	// Limit is too large
	err = checkMaxLimit(limit)
	if err != nil {
		http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
		return
	}

	// Get the value descriptors
	vdList, err := getValueDescriptorsByLabel(label)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var vdNames []string
	for _, vd := range vdList {
		vdNames = append(vdNames, vd.Name)
	}

	readings, err := getReadingsByValueDescriptorNames(vdNames, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error escaping the type: " + err.Error())
		return
	}

	limit, err := strconv.Atoi(vars["limit"])
	// Problem converting to int
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error converting the limit to an integer: " + err.Error())
		return
	}

	err = checkMaxLimit(limit)
	if err != nil {
		http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
		return
	}

	// Get the value descriptors
	vdList, err := getValueDescriptorsByType(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
		return
	}
	var vdNames []string
	for _, vd := range vdList {
		vdNames = append(vdNames, vd.Name)
	}

	readings, err := getReadingsByValueDescriptorNames(vdNames, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
		return
	}

	encode(readings, w)
}

// Return a list of readings between the start and end (creation time)
// /reading/{start}/{end}/{limit}
func readingByCreationTimeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	start, err := strconv.ParseInt((vars["start"]), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error converting the start time to an integer: " + err.Error())
		return
	}
	end, err := strconv.ParseInt((vars["end"]), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error converting the end time to an integer: " + err.Error())
		return
	}
	limit, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error converting the limit to an integer: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		err = checkMaxLimit(limit)
		if err != nil {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			return
		}

		readings, err := getReadingsByCreationTime(start, end, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error unescaping the value descriptor name: " + err.Error())
		return
	}

	device, err := url.QueryUnescape(vars["device"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error unescaping the device: " + err.Error())
		return
	}

	limit, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error converting limit to an integer: " + err.Error())
		return
	}

	err = checkMaxLimit(limit)
	if err != nil {
		http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
		return
	}

	// Check device
	if err := checkDevice(device); err != nil {
		LoggingClient.Error(fmt.Sprintf("error checking device %s %v", device, err))
		switch err := err.(type) {
		case *types.ErrServiceClient:
			http.Error(w, err.Error(), err.StatusCode)
			return
		default: //return an error on everything else.
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
	}

	// Check for value descriptor
	if Configuration.Writable.ValidateCheck {
		_, err = getValueDescriptorByName(name)
		if err != nil {
			switch err.(type) {
			case *errors.ErrDbNotFound:
				http.Error(w, "Value descriptor not found for reading", http.StatusConflict)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	readings, err := getReadingsByDeviceAndValueDescriptor(device, name, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	encode(readings, w)
}

// Value Descriptors

// GET, POST, and PUT for value descriptors
// api/v1/valuedescriptor
func valueDescriptorHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case http.MethodGet:
		vList, err := getAllValueDescriptors()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Check the limit
		err = checkMaxLimit(len(vList))
		if err != nil {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			return
		}

		encode(vList, w)
	case http.MethodPost:
		v, err := decodeValueDescriptor(r.Body)
		if err != nil {
			switch err.(type) {
			case *errors.ErrJsonDecoding:
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			case *errors.ErrValueDescriptorInvalid:
				http.Error(w, err.Error(), http.StatusConflict)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		id, err := addValueDescriptor(v)
		if err != nil {
			switch err.(type) {
			case *errors.ErrValueDescriptorInUse:
				http.Error(w, err.Error(), http.StatusConflict)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(id))
	case http.MethodPut:
		vd, err := decodeValueDescriptor(r.Body)
		if err != nil {
			switch err.(type) {
			case *errors.ErrJsonDecoding:
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			case *errors.ErrValueDescriptorInvalid:
				http.Error(w, err.Error(), http.StatusConflict)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		err = updateValueDescriptor(vd)
		if err != nil {
			switch err.(type) {
			case *errors.ErrDbNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			case *errors.ErrValueDescriptorInvalid:
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			case *errors.ErrValueDescriptorInUse:
				http.Error(w, "Data integrity issue. Value Descriptor still in use by readings", http.StatusConflict)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
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

	err := deleteValueDescriptorById(id)
	if err != nil {
		switch err.(type) {
		case *errors.ErrDbNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		case *errors.ErrValueDescriptorInvalid:
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		case *errors.ErrValueDescriptorInUse:
			http.Error(w, "Data integrity issue. Value Descriptor still in use by readings", http.StatusConflict)
			return
		case *errors.ErrInvalidId:
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Value descriptors based on name
// api/v1/valuedescriptor/name/{name}
func valueDescriptorByNameHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	name, err := url.QueryUnescape(vars["name"])

	// Problems unescaping
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error unescaping the value descriptor name: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		v, err := dbClient.ValueDescriptorByName(name)
		if err != nil {
			switch err := err.(type) {
			case *types.ErrServiceClient:
				http.Error(w, err.Error(), err.StatusCode)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}
		encode(v, w)
	case http.MethodDelete:
		if err = deleteValueDescriptorByName(name); err != nil {
			switch err.(type) {
			case *errors.ErrDbNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			case *errors.ErrValueDescriptorInvalid:
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			case *errors.ErrValueDescriptorInUse:
				http.Error(w, "Data integrity issue. Value Descriptor still in use by readings", http.StatusConflict)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
	}
}

// Get a value descriptor based on the ID
// HTTP 404 not found if the ID isn't in the database
// api/v1/valuedescriptor/{id}
func valueDescriptorByIdHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	id := vars["id"]

	switch r.Method {
	case http.MethodGet:
		vd, err := getValueDescriptorById(id)
		if err != nil {
			switch err.(type) {
			case *errors.ErrDbNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		encode(vd, w)
	}
}

// Get the value descriptor from the UOM label
// api/v1/valuedescriptor/uomlabel/{uomLabel}
func valueDescriptorByUomLabelHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	uomLabel, err := url.QueryUnescape(vars["uomLabel"])

	// Problem unescaping
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error unescaping the UOM Label of the value descriptor: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		vdList, err := getValueDescriptorsByUomLabel(uomLabel)
		if err != nil {
			switch err.(type) {
			case *errors.ErrDbNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		encode(vdList, w)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error unescaping label for the value descriptor: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		v, err := getValueDescriptorsByLabel(label)
		if err != nil {
			switch err.(type) {
			case *errors.ErrDbNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		encode(v, w)
	}
}

// Return the value descriptors that are associated with a device
// The value descriptor is expected parameters on puts or expected values on get/put commands
// api/v1/valuedescriptor/devicename/{device}
func valueDescriptorByDeviceHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)

	device, err := url.QueryUnescape(vars["device"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error unescaping the device: " + err.Error())
		return
	}

	// Get the value descriptors
	vdList, err := getValueDescriptorsByDeviceName(device)
	if err != nil {
		switch err := err.(type) {
		case *errors.ErrDbNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		case *types.ErrServiceClient:
			http.Error(w, err.Error(), err.StatusCode)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error unescaping the device ID: " + err.Error())
		return
	}

	// Get the value descriptors
	vdList, err := getValueDescriptorsByDeviceId(deviceId)
	if err != nil {
		switch err := err.(type) {
		case *types.ErrServiceClient:
			http.Error(w, err.Error(), err.StatusCode)
		case *errors.ErrDbNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	encode(vdList, w)
}

func metricsHandler(w http.ResponseWriter, _ *http.Request) {
	var t internal.Telemetry

	// The micro-service is to be considered the System Of Record (SOR) in terms of accurate information.
	// Fetch metrics for the data service.
	var rtm runtime.MemStats

	// Read full memory stats
	runtime.ReadMemStats(&rtm)

	// Miscellaneous memory stats
	t.Alloc = rtm.Alloc
	t.TotalAlloc = rtm.TotalAlloc
	t.Sys = rtm.Sys
	t.Mallocs = rtm.Mallocs
	t.Frees = rtm.Frees

	// Live objects = Mallocs - Frees
	t.LiveObjects = t.Mallocs - t.Frees

	encode(t, w)

	return
}
