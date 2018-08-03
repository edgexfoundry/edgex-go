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

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
)

// Reading handler
// GET, PUT, and POST readings
func readingHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case http.MethodGet:
		r, err := dbClient.Readings()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
			return
		}

		// Check max limit
		if len(r) > configuration.ReadMaxLimit {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			LoggingClient.Error(maxExceededString)
			return
		}

		encode(r, w)
	case http.MethodPost:
		reading := models.Reading{}
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&reading)

		// Problem decoding
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			LoggingClient.Error("Error decoding the reading: " + err.Error())
			return
		}

		if configuration.ValidateCheck {
			// Check the value descriptor
			vd, err := dbClient.ValueDescriptorByName(reading.Name)
			if err != nil {
				if err == db.ErrNotFound {
					http.Error(w, "Value descriptor not found for reading", http.StatusConflict)
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				LoggingClient.Error(err.Error())
				return
			}

			valid, err := isValidValueDescriptor(vd, reading)
			if !valid {
				http.Error(w, "Validation failed", http.StatusConflict)
				LoggingClient.Error("Validation failed")
				return
			}
		}

		// Check device
		if reading.Device != "" {
			if checkDevice(reading.Device, w) == false {
				return
			}
		}

		if configuration.PersistData {
			id, err := dbClient.AddReading(reading)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				LoggingClient.Error(err.Error())
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
			http.Error(w, err.Error(), http.StatusBadRequest)
			LoggingClient.Error("Error decoding the reading: " + err.Error())
			return
		}

		// Check if the reading exists
		to, err := dbClient.ReadingById(from.Id.Hex())
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Reading not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}

		// Update the fields
		if from.Value != "" {
			to.Value = from.Value
		}
		if from.Name != "" {
			to.Name = from.Name
		}
		if from.Origin != 0 {
			to.Origin = from.Origin
		}

		if from.Value != "" || from.Name != "" {
			if configuration.ValidateCheck {
				// Check the value descriptor
				vd, err := dbClient.ValueDescriptorByName(to.Name)
				if err != nil {
					if err == db.ErrNotFound {
						http.Error(w, "Value descriptor not found for reading", http.StatusConflict)
					} else {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
					LoggingClient.Error(err.Error())
					return
				}

				fmt.Println(to)

				valid, err := isValidValueDescriptor(vd, to)
				if !valid {
					http.Error(w, "Validation failed", http.StatusConflict)
					LoggingClient.Error("Validation failed")
					return
				}
			}
		}

		err = dbClient.UpdateReading(to)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
			return
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
		reading, err := dbClient.ReadingById(id)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Reading not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
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
		count, err := dbClient.ReadingCount()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(strconv.Itoa(count)))
		if err != nil {
			LoggingClient.Error(err.Error(), "")
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
		reading, err := dbClient.ReadingById(id)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Reading not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}

		err = dbClient.DeleteReadingById(reading.Id.Hex())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
			return
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
		if limit > configuration.ReadMaxLimit {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			LoggingClient.Error(maxExceededString)
			return
		}

		// Check device
		if checkDevice(deviceId, w) == false {
			return
		}

		readings, err := dbClient.ReadingsByDevice(deviceId, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
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

	// Check for value descriptor
	if configuration.ValidateCheck {
		_, err = dbClient.ValueDescriptorByName(name)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Value descriptor not found for reading", http.StatusConflict)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}
	}

	// Limit is too large
	if limit > configuration.ReadMaxLimit {
		http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
		LoggingClient.Error(maxExceededString)
		return
	}

	read, err := dbClient.ReadingsByValueDescriptor(name, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
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
	if limit > configuration.ReadMaxLimit {
		http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
		LoggingClient.Error(maxExceededString)
		return
	}

	// Get the value descriptors
	vList, err := dbClient.ValueDescriptorsByUomLabel(uomLabel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
		return
	}

	var vNames []string
	for _, v := range vList {
		vNames = append(vNames, v.Name)
	}

	readings, err := dbClient.ReadingsByValueDescriptorNames(vNames, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
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
	if limit > configuration.ReadMaxLimit {
		LoggingClient.Error(maxExceededString)
		http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
		return
	}

	// Get the value descriptors
	vdList, err := dbClient.ValueDescriptorsByLabel(label)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
		return
	}
	var vdNames []string
	for _, vd := range vdList {
		vdNames = append(vdNames, vd.Name)
	}

	readings, err := dbClient.ReadingsByValueDescriptorNames(vdNames, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
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

	l, err := strconv.Atoi(vars["limit"])
	// Problem converting to int
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error converting the limit to an integer: " + err.Error())
		return
	}

	// Limit exceeds max limit
	if l > configuration.ReadMaxLimit {
		http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
		LoggingClient.Error(maxExceededString)
		return
	}

	// Get the value descriptors
	vdList, err := dbClient.ValueDescriptorsByType(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
		return
	}
	var vdNames []string
	for _, vd := range vdList {
		vdNames = append(vdNames, vd.Name)
	}

	readings, err := dbClient.ReadingsByValueDescriptorNames(vdNames, l)
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
	s, err := strconv.ParseInt((vars["start"]), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error converting the start time to an integer: " + err.Error())
		return
	}
	e, err := strconv.ParseInt((vars["end"]), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error converting the end time to an integer: " + err.Error())
		return
	}
	l, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error converting the limit to an integer: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		if l > configuration.ReadMaxLimit {
			LoggingClient.Error(maxExceededString)
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			return
		}

		readings, err := dbClient.ReadingsByCreationTime(s, e, l)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
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

	if limit > configuration.ReadMaxLimit {
		LoggingClient.Error(maxExceededString)
		http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
		return
	}

	// Check device
	if checkDevice(device, w) == false {
		return
	}

	// Check for value descriptor
	if configuration.ValidateCheck {
		_, err = dbClient.ValueDescriptorByName(name)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Value descriptor not found for reading", http.StatusConflict)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}
	}

	readings, err := dbClient.ReadingsByDeviceAndValueDescriptor(device, name, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
		return
	}

	encode(readings, w)
}
