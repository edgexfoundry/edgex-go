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
	"regexp"

	"github.com/edgexfoundry/edgex-go/core/clients/metadataclients"
	"github.com/edgexfoundry/edgex-go/core/data/clients"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/gorilla/mux"
)

const (
	formatSpecifier          = "%(\\d+\\$)?([-#+ 0,(\\<]*)?(\\d+)?(\\.\\d+)?([tT])?([a-zA-Z%])"
	maxExceededString string = "Error, exceeded the max limit as defined in config"
)

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
	case http.MethodGet:
		vList, err := dbc.ValueDescriptors()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}

		// Check the limit
		if len(vList) > configuration.ReadMaxLimit {
			http.Error(w, maxExceededString, http.StatusRequestEntityTooLarge)
			loggingClient.Error(maxExceededString)
			return
		}

		encode(vList, w)
	case http.MethodPost:
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
	case http.MethodPut:
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
		//encode(true, w)
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
	//encode(true, w)
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
	case http.MethodGet:
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
	case http.MethodDelete:
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
		//encode(true, w)
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
	case http.MethodGet:
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
	case http.MethodGet:
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
	case http.MethodGet:
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
