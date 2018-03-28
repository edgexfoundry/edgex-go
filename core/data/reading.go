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
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/edgexfoundry/edgex-go/core/data/clients"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/gorilla/mux"
)

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

// Reading handler
// GET, PUT, and POST readings
func readingHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case http.MethodGet:
		r, err := dbc.Readings()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}

		// Check max limit
		if len(r) > configuration.ReadMaxLimit {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			loggingClient.Error(maxExceededString)
			return
		}

		encode(r, w)
	case http.MethodPost:
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

		if configuration.PersistData {
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
	case http.MethodPut:
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
		//encode(true, w)
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
	case http.MethodGet:
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
	case http.MethodDelete:
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
		//encode(true, w)
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
	case http.MethodGet:
		if limit > configuration.ReadMaxLimit {
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
				if configuration.MetaDataCheck {
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
	if limit > configuration.ReadMaxLimit {
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
	if limit > configuration.ReadMaxLimit {
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
	if limit > configuration.ReadMaxLimit {
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
	if l > configuration.ReadMaxLimit {
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
	case http.MethodGet:
		if l > configuration.ReadMaxLimit {
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

	if limit > configuration.ReadMaxLimit {
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
