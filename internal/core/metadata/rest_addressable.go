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
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	types "github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
)

func restGetAllAddressables(w http.ResponseWriter, _ *http.Request) {
	results, err := getAllAddressables()
	if err != nil {
		switch err.(type) {
		case *types.ErrLimitExceeded:
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Add a new addressable
// The name must be unique
func restAddAddressable(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var a models.Addressable
	err := json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := addAddressable(a)
	if err != nil {
		switch err.(type) {
		case *types.ErrDuplicateAddressableName:
			http.Error(w, err.Error(), http.StatusConflict)
		case *types.ErrEmptyAddressableName:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(id))
	if err != nil {
		LoggingClient.Error(err.Error())
		return
	}
}

// Update addressable by ID or name (ID used first)
// 404 Not found if no addressable is found to update
func restUpdateAddressable(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var ra models.Addressable
	if err := json.NewDecoder(r.Body).Decode(&ra); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := updateAddressable(ra); err != nil {
		switch err.(type) {
		case *types.ErrAddressableNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case *types.ErrAddressableInUse:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("true")); err != nil {
		LoggingClient.Error(err.Error())
		return
	}
}

func restGetAddressableById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var id string = vars["id"]
	result, err := dbClient.GetAddressableById(id)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
func restDeleteAddressableById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var id string = vars[ID]

	// Check if the addressable exists
	a, err := dbClient.GetAddressableById(id)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Check if the addressable is still in use
	isStillInUse, err := isAddressableStillInUse(a)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if isStillInUse {
		err = errors.New("Data integrity issue: attempt to delete addressable still in use by a device or device service")
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	err = dbClient.DeleteAddressableById(a.Id)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("true"))
}
func restDeleteAddressableByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	// Problems unescaping
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the addressable exists
	a, err := dbClient.GetAddressableByName(n)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Check if the addressable is still in use
	isStillInUse, err := isAddressableStillInUse(a)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	if isStillInUse {
		err = errors.New("Data integrity issue: attempt to delete addressable still in use by a device or device service")
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	if err := dbClient.DeleteAddressableById(a.Id); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Helper function to determine if an addressable is still referenced by a device or device service
func isAddressableStillInUse(a models.Addressable) (bool, error) {
	// Check devices
	var d []models.Device
	err := dbClient.GetDevicesByAddressableId(&d, a.Id)
	if err != nil {
		return false, err
	}
	if len(d) > 0 {
		return true, nil
	}

	// Check device services
	ds, err := dbClient.GetDeviceServicesByAddressableId(a.Id)
	if err != nil {
		return false, err
	}
	if len(ds) > 0 {
		return true, nil
	}

	return false, nil
}
func restGetAddressableByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dn, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	result, err := dbClient.GetAddressableByName(dn)
	if err != nil {
		LoggingClient.Error(err.Error())
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func restGetAddressableByTopic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	t, err := url.QueryUnescape(vars[TOPIC])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := dbClient.GetAddressablesByTopic(t)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
func restGetAddressableByPort(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var strp string = vars[PORT]
	p, err := strconv.Atoi(strp)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := dbClient.GetAddressablesByPort(p)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
func restGetAddressableByPublisher(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p, err := url.QueryUnescape(vars[PUBLISHER])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := dbClient.GetAddressablesByPublisher(p)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
func restGetAddressableByAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	a, err := url.QueryUnescape(vars[ADDRESS])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := dbClient.GetAddressablesByAddress(a)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Notify the associated device services for changes in the device addressable
func notifyAddressableAssociates(a models.Addressable, action string) error {
	// Get the devices
	var d []models.Device
	if err := dbClient.GetDevicesByAddressableId(&d, a.Id); err != nil {
		LoggingClient.Error(err.Error())
		return err
	}

	// Get the services for each device
	// Use map as a Set
	dsMap := map[string]models.DeviceService{}
	var ds []models.DeviceService
	for _, device := range d {
		// Only add if not there
		if _, ok := dsMap[device.Service.Service.Id]; !ok {
			dsMap[device.Service.Service.Id] = device.Service
			ds = append(ds, device.Service)
		}
	}

	if err := notifyAssociates(ds, a.Id, action, models.ADDRESSABLE); err != nil {
		LoggingClient.Error(err.Error())
		return err
	}

	return nil
}
