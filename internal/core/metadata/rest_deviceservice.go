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
package metadata

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
)

// Get the addressable by its ID or Name
func getAddressableByIdOrName(a *models.Addressable, w http.ResponseWriter) error {
	id := a.Id
	name := a.Name

	// Try by ID
	if _, err := dbClient.GetAddressableById(id); err != nil {
		// Try by name
		if _, err = dbClient.GetAddressableByName(name); err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Addressable not found", http.StatusServiceUnavailable)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			return err
		}
	}

	return nil
}

func restGetAllDeviceServices(w http.ResponseWriter, _ *http.Request) {
	r := make([]models.DeviceService, 0)
	if err := dbClient.GetAllDeviceServices(&r); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check the limit
	if len(r) > Configuration.Service.ReadMaxLimit {
		err := errors.New("Max limit exceeded")
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&r)
}

func restAddDeviceService(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var ds models.DeviceService
	err := json.NewDecoder(r.Body).Decode(&ds)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Addressable Check
	// No ID or Name given for addressable
	if ds.Service.Addressable.Id == "" && ds.Service.Addressable.Name == "" {
		err = errors.New("Must provide an Addressable for Device Service")
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	var addressable models.Addressable
	// First try by name
	addressable, err = dbClient.GetAddressableByName(ds.Service.Addressable.Name)
	if err != nil && err == db.ErrNotFound {
		addressable, err = dbClient.GetAddressableById(ds.Service.Addressable.Id)
	}
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Addressable not found by ID or Name", http.StatusNotFound)
			LoggingClient.Error("Addressable not found by ID or Name: "+err.Error(), "")
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
		}
		return
	}
	ds.Service.Addressable = addressable

	// Add the device service
	if err := dbClient.AddDeviceService(&ds); err != nil {
		if err == db.ErrNotUnique {
			http.Error(w, "Duplicate name for the device service", http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(ds.Service.Id.Hex()))
}

// Get all the addressables for the devices that are associated with the device service
func restGetAddressablesForAssociatedDevicesById(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	var id string = vars[ID]
	var ds models.DeviceService

	// Check if the device service exists
	if err := dbClient.GetDeviceServiceById(&ds, id); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Device service not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		LoggingClient.Error(err.Error())
		return
	}

	addressables := []models.Addressable{}

	if err := getAddressablesForAssociatedDevices(&addressables, ds, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(addressables)
}

// Get all the addressables fo the devices that are associated with the device service
func restGetAddressablesForAssociatedDevicesByName(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device service exists
	var ds models.DeviceService
	if err = dbClient.GetDeviceServiceByName(&ds, n); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Device service not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		LoggingClient.Error(err.Error())
		return
	}

	addressables := []models.Addressable{}
	if err = getAddressablesForAssociatedDevices(&addressables, ds, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(addressables)
}

// Get the addressables for the associated devices to the device service
// addressables will have the result
func getAddressablesForAssociatedDevices(addressables *[]models.Addressable, ds models.DeviceService, w http.ResponseWriter) error {
	// Get the associated devices
	var devices []models.Device
	if err := dbClient.GetDevicesByServiceId(&devices, ds.Service.Id.Hex()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	// Get the addressables for all the devices
	// Use a map to maintain a set (no duplicates)
	// Convert to a slice afterwards
	aMap := map[string]models.Addressable{}
	for _, d := range devices {
		// Only append addressable if its not in the map
		if _, ok := aMap[d.Addressable.Id]; !ok {
			aMap[d.Addressable.Id] = d.Addressable
			*addressables = append(*addressables, d.Addressable)
		}
	}

	return nil
}

func restUpdateDeviceService(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var from models.DeviceService
	err := json.NewDecoder(r.Body).Decode(&from)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the device service exists and get it
	var to models.DeviceService
	// Try by ID
	if err = dbClient.GetDeviceServiceById(&to, from.Service.Id.Hex()); err != nil {
		// Try by Name
		if err = dbClient.GetDeviceServiceByName(&to, from.Service.Name); err != nil {
			http.Error(w, "Device service not found", http.StatusNotFound)
			LoggingClient.Error(err.Error())
			return
		}
	}

	if err = updateDeviceServiceFields(from, &to, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}

	if err := dbClient.UpdateDeviceService(to); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the relevant device service fields
func updateDeviceServiceFields(from models.DeviceService, to *models.DeviceService, w http.ResponseWriter) error {
	// Use .String() to compare empty structs (not ideal, but there is no .equals method)
	if (from.Service.Addressable.String() != models.Addressable{}.String()) {
		// Check if addressable exists
		to.Service.Addressable = from.Service.Addressable
		if err := getAddressableByIdOrName(&to.Service.Addressable, w); err != nil {
			return err
		}
	}
	if from.AdminState != models.AdminState("") {
		if !models.IsAdminStateType(string(from.AdminState)) {
			err := errors.New("Invalid Admin State: " + string(from.AdminState) + " Must be 'locked' or 'unlocked'")
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return err
		}
		to.AdminState = from.AdminState
	}
	if from.Service.Description != "" {
		to.Service.Description = from.Service.Description
	}
	if from.Service.Labels != nil {
		to.Service.Labels = from.Service.Labels
	}
	if from.Service.LastConnected != 0 {
		to.Service.LastConnected = from.Service.LastConnected
	}
	if from.Service.LastReported != 0 {
		to.Service.LastReported = from.Service.LastReported
	}
	if from.Service.Name != "" {
		to.Service.Name = from.Service.Name

		// Check if the new name is unique
		var checkDS models.DeviceService
		err := dbClient.GetDeviceServiceByName(&checkDS, from.Service.Name)
		if err != nil {
			// A problem occurred accessing database
			if err != db.ErrNotFound {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return err
			}
		}

		// Found a device service, make sure its the one we're trying to update
		if err != db.ErrNotFound {
			// Different IDs -> Name is not unique
			if checkDS.Service.Id != to.Service.Id {
				err = errors.New("Duplicate name for Device Service")
				http.Error(w, err.Error(), http.StatusConflict)
				return err
			}
		}
	}
	if from.Service.OperatingState != models.OperatingState("") {
		// Check operating state
		if !models.IsOperatingStateType(string(from.Service.OperatingState)) {
			err := errors.New("Invalid operating state: " + string(from.Service.OperatingState) + " Must be 'enabled' or 'disabled'")
			http.Error(w, err.Error(), http.StatusConflict)
			return err
		}

		to.Service.OperatingState = from.Service.OperatingState
	}
	if from.Service.Origin != 0 {
		to.Service.Origin = from.Service.Origin
	}

	return nil
}

func restGetServiceByAddressableName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	an, err := url.QueryUnescape(vars[ADDRESSABLENAME])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res := make([]models.DeviceService, 0)

	// Check if the addressable exists
	a, err := dbClient.GetAddressableByName(an)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Addressable not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	if err = dbClient.GetDeviceServicesByAddressableId(&res, a.Id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetServiceByAddressableId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var sid string = vars[ADDRESSABLEID]
	res := make([]models.DeviceService, 0)

	// Check if the Addressable exists
	_, err := dbClient.GetAddressableById(sid)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Addressable not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	if err := dbClient.GetDeviceServicesByAddressableId(&res, sid); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetServiceWithLabel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	l, err := url.QueryUnescape(vars[LABEL])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res := make([]models.DeviceService, 0)

	if err := dbClient.GetDeviceServicesWithLabel(&res, l); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetServiceByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dn, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var res models.DeviceService
	err = dbClient.GetDeviceServiceByName(&res, dn)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restDeleteServiceById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var id string = vars[ID]

	// Check if the device service exists and get it
	var ds models.DeviceService
	if err := dbClient.GetDeviceServiceById(&ds, id); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Device service not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	if err := deleteDeviceService(ds, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("true"))
}

func restDeleteServiceByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the device service exists
	var ds models.DeviceService
	if err = dbClient.GetDeviceServiceByName(&ds, n); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Device service not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	// Delete the device service
	if err = deleteDeviceService(ds, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Delete the device service
// Delete the associated devices
// Delete the associated provision watchers
func deleteDeviceService(ds models.DeviceService, w http.ResponseWriter) error {
	// Delete the associated devices
	var devices []models.Device
	if err := dbClient.GetDevicesByServiceId(&devices, ds.Service.Id.Hex()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}
	for _, device := range devices {
		if err := deleteDevice(device, w); err != nil {
			return err
		}
	}

	// Delete the associated provision watchers
	var watchers []models.ProvisionWatcher
	if err := dbClient.GetProvisionWatchersByServiceId(&watchers, ds.Service.Id.Hex()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}
	for _, watcher := range watchers {
		if err := deleteProvisionWatcher(watcher, w); err != nil {
			return err
		}
	}

	// Delete the device service
	if err := dbClient.DeleteDeviceServiceById(ds.Id.Hex()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	return nil
}

func restUpdateServiceLastConnectedById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var id string = vars[ID]
	var vlc string = vars[LASTCONNECTED]
	lc, err := strconv.ParseInt(vlc, 10, 64)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the device service exists
	var ds models.DeviceService
	if err = dbClient.GetDeviceServiceById(&ds, id); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Device service not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	if err = updateServiceLastConnected(ds, lc, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restUpdateServiceLastConnectedByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var vlc string = vars[LASTCONNECTED]
	lc, err := strconv.ParseInt(vlc, 10, 64)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the device service exists
	var ds models.DeviceService
	if err = dbClient.GetDeviceServiceByName(&ds, n); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Device service not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	// Update last connected
	if err = updateServiceLastConnected(ds, lc, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the last connected value of the device service
func updateServiceLastConnected(ds models.DeviceService, lc int64, w http.ResponseWriter) error {
	ds.LastConnected = lc

	if err := dbClient.UpdateDeviceService(ds); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	return nil
}

func restGetServiceById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars[ID]
	var res models.DeviceService

	if err := dbClient.GetDeviceServiceById(&res, did); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restUpdateServiceOpStateById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var id string = vars[ID]
	var os string = vars[OPSTATE]

	// Check the OpState
	newOs, f := models.GetOperatingState(os)
	if !f {
		err := errors.New("Invalid State: " + os + " Must be 'ENABLED' or 'DISABLED'")
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the device service exists
	var ds models.DeviceService
	if err := dbClient.GetDeviceServiceById(&ds, id); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Device service not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	if err := updateServiceOpState(ds, newOs, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restUpdateServiceOpStateByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var os string = vars[OPSTATE]

	// Check the OpState
	newOs, f := models.GetOperatingState(os)
	if !f {
		err = errors.New("Invalid State: " + os + " Must be 'ENABLED' or 'DISABLED'")
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the device service exists
	var ds models.DeviceService
	if err = dbClient.GetDeviceServiceByName(&ds, n); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	if err := updateServiceOpState(ds, newOs, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the OpState for the device service
func updateServiceOpState(ds models.DeviceService, os models.OperatingState, w http.ResponseWriter) error {
	ds.OperatingState = os
	if err := dbClient.UpdateDeviceService(ds); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	return nil
}

func restUpdateServiceAdminStateById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var id string = vars[ID]
	var as string = vars[ADMINSTATE]

	// Check the admin state
	newAs, f := models.GetAdminState(as)
	if !f {
		err := errors.New("Invalid state: " + as + " Must be 'LOCKED' or 'UNLOCKED'")
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	// Check if the device service exists
	var ds models.DeviceService
	if err := dbClient.GetDeviceServiceById(&ds, id); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Device service not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	// Update the admin state
	if err := updateServiceAdminState(ds, newAs, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restUpdateServiceAdminStateByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var as string = vars[ADMINSTATE]

	// Check the admin state
	newAs, f := models.GetAdminState(as)
	if !f {
		err := errors.New("Invalid state: " + as + " Must be 'LOCKED' or 'UNLOCKED'")
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	// Check if the device service exists
	var ds models.DeviceService
	if err = dbClient.GetDeviceServiceByName(&ds, n); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Device service not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	// Update the admins state
	if err = updateServiceAdminState(ds, newAs, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the admin state for the device service
func updateServiceAdminState(ds models.DeviceService, as models.AdminState, w http.ResponseWriter) error {
	ds.AdminState = as
	if err := dbClient.UpdateDeviceService(ds); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	return nil
}

func restUpdateServiceLastReportedById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var id string = vars[ID]
	var vlr string = vars[LASTREPORTED]
	lr, err := strconv.ParseInt(vlr, 10, 64)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the devicde service exists
	var ds models.DeviceService
	if err = dbClient.GetDeviceServiceById(&ds, id); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Device service not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	if err = updateServiceLastReported(ds, lr, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restUpdateServiceLastReportedByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var vlr string = vars[LASTREPORTED]
	lr, err := strconv.ParseInt(vlr, 10, 64)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the device service exists
	var ds models.DeviceService
	if err = dbClient.GetDeviceServiceByName(&ds, n); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Device service not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	if err = updateServiceLastReported(ds, lr, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the last reported value for the device service
func updateServiceLastReported(ds models.DeviceService, lr int64, w http.ResponseWriter) error {
	ds.LastReported = lr
	if err := dbClient.UpdateDeviceService(ds); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	return nil
}

// Notify associates (associated device services)
// This function is called when an object changes in metadata
func notifyAssociates(deviceServices []models.DeviceService, id string, action string, actionType models.ActionType) error {
	for _, ds := range deviceServices {
		if err := callback(ds, id, action, actionType); err != nil {
			return err
		}
	}

	return nil
}

// Make the callback for the device service
func callback(service models.DeviceService, id string, action string, actionType models.ActionType) error {
	client := &http.Client{}
	url := service.Service.Addressable.GetCallbackURL()
	if len(url) > 0 {
		body, err := getBody(id, actionType)
		if err != nil {
			return err
		}
		req, err := http.NewRequest(string(action), url, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Add("Content-Type", "application/json")

		go makeRequest(client, req)
	} else {
		LoggingClient.Info("callback::no addressable for " + service.Name)
	}
	return nil
}

// Asynchronous call
func makeRequest(client *http.Client, req *http.Request) {
	// Make the request
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		resp.Close = true
	} else {
		LoggingClient.Error(err.Error())
	}
}

// Turn the ID and ActionType into the JSON body that will be passed
func getBody(id string, actionType models.ActionType) ([]byte, error) {
	return json.Marshal(models.CallbackAlert{ActionType: actionType, Id: id})
}
