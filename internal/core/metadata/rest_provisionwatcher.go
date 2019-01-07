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

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
)

func restGetProvisionWatchers(w http.ResponseWriter, _ *http.Request) {
	res, err := dbClient.GetAllProvisionWatchers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		LoggingClient.Error(err.Error())
		return
	}

	// Check the length
	if len(res) > Configuration.Service.ReadMaxLimit {
		err := errors.New("Max limit exceeded")
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&res)
}

func restDeleteProvisionWatcherById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var id string = vars[ID]

	// Check if the provision watcher exists
	pw, err := dbClient.GetProvisionWatcherById(id)
	if err != nil {
		errMessage := "Provision Watcher not found by ID: " + err.Error()
		LoggingClient.Error(errMessage)
		http.Error(w, errMessage, http.StatusNotFound)
		return
	}

	err = deleteProvisionWatcher(pw, w)
	if err != nil {
		errMessage := "Error deleting provision watcher"
		LoggingClient.Error(errMessage)
		http.Error(w, errMessage, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("true"))
}

func restDeleteProvisionWatcherByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	// Check if the provision watcher exists
	pw, err := dbClient.GetProvisionWatcherByName(n)
	if err != nil {
		if err == db.ErrNotFound {
			errMessage := "Provision watcher not found: " + err.Error()
			http.Error(w, errMessage, http.StatusNotFound)
			LoggingClient.Error(errMessage)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	if err = deleteProvisionWatcher(pw, w); err != nil {
		LoggingClient.Error("Problem deleting provision watcher: " + err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Delete the provision watcher
func deleteProvisionWatcher(pw models.ProvisionWatcher, w http.ResponseWriter) error {
	if err := dbClient.DeleteProvisionWatcherById(pw.Id); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	if err := notifyProvisionWatcherAssociates(pw, http.MethodDelete); err != nil {
		LoggingClient.Error("Problem notifying associated device services to provision watcher: " + err.Error())
	}

	return nil
}

func restGetProvisionWatcherById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var id string = vars[ID]

	res, err := dbClient.GetProvisionWatcherById(id)
	if err != nil {
		if err == db.ErrNotFound {
			errMessage := "Problem getting provision watcher by ID: " + err.Error()
			LoggingClient.Error(errMessage)
			http.Error(w, errMessage, http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetProvisionWatcherByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	res, err := dbClient.GetProvisionWatcherByName(n)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			LoggingClient.Error("Provision watcher not found: " + err.Error())
		} else {
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetProvisionWatchersByProfileId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var pid string = vars[ID]

	// Check if the device profile exists
	if _, err := dbClient.GetDeviceProfileById(pid); err != nil {
		LoggingClient.Error("Device profile not found: " + err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	res, err := dbClient.GetProvisionWatchersByProfileId(pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Problem getting provision watcher: " + err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetProvisionWatchersByProfileName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pn, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	// Check if the device profile exists
	dp, err := dbClient.GetDeviceProfileByName(pn)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Device profile not found", http.StatusNotFound)
			LoggingClient.Error("Device profile not found: " + err.Error())
		} else {
			LoggingClient.Error("Problem getting device profile: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	res, err := dbClient.GetProvisionWatchersByProfileId(dp.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Problem getting provision watcher: " + err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetProvisionWatchersByServiceId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var sid string = vars[ID]

	// Check if the device service exists
	if _, err := dbClient.GetDeviceServiceById(sid); err != nil {
		http.Error(w, "Device Service not found", http.StatusNotFound)
		LoggingClient.Error("Device service not found: " + err.Error())
		return
	}

	res, err := dbClient.GetProvisionWatchersByServiceId(sid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Problem getting provision watcher: " + err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetProvisionWatchersByServiceName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sn, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	// Check if the device service exists
	ds, err := dbClient.GetDeviceServiceByName(sn)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Device service not found", http.StatusNotFound)
			LoggingClient.Error("Device service not found: " + err.Error())
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error("Problem getting device service: " + err.Error())
		}
		return
	}

	// Get the provision watchers
	res, err := dbClient.GetProvisionWatchersByServiceId(ds.Service.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		LoggingClient.Error("Problem getting provision watcher: " + err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetProvisionWatchersByIdentifier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	k, err := url.QueryUnescape(vars[KEY])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}
	v, err := url.QueryUnescape(vars[VALUE])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	res, err := dbClient.GetProvisionWatchersByIdentifier(k, v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Problem getting provision watchers: " + err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restAddProvisionWatcher(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var pw models.ProvisionWatcher
	var err error

	if err = json.NewDecoder(r.Body).Decode(&pw); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the name exists
	if pw.Name == "" {
		err = errors.New("No name provided for new provision watcher")
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// Check if the device profile exists
	// Try by ID
	var profile models.DeviceProfile
	if pw.Profile.Id != "" {
		profile, err = dbClient.GetDeviceProfileById(pw.Profile.Id)
	}
	if pw.Profile.Id == "" || err != nil {
		// Try by name
		if profile, err = dbClient.GetDeviceProfileByName(pw.Profile.Name); err != nil {
			if err == db.ErrNotFound {
				LoggingClient.Error("Device profile not found for provision watcher: " + err.Error())
				http.Error(w, "Device profile not found for provision watcher", http.StatusConflict)
			} else {
				LoggingClient.Error("Problem getting device profile: " + err.Error())
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			return
		}
	}
	pw.Profile = profile

	// Check if the device service exists
	// Try by ID
	var service models.DeviceService
	if pw.Service.Service.Id != "" {
		service, err = dbClient.GetDeviceServiceById(pw.Service.Service.Id)
	}
	if pw.Service.Service.Id == "" || err != nil {
		// Try by name
		if service, err = dbClient.GetDeviceServiceByName(pw.Service.Service.Name); err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Device service not found for provision watcher", http.StatusConflict)
				LoggingClient.Error("Device service not found for provision watcher: " + err.Error())
			} else {
				LoggingClient.Error("Problem getting device service for provision watcher: " + err.Error())
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			return
		}
	}
	pw.Service = service

	id, err := dbClient.AddProvisionWatcher(pw)
	if err != nil {
		if err == db.ErrNotUnique {
			LoggingClient.Error("Duplicate name for the provision watcher: " + err.Error())
			http.Error(w, "Duplicate name for the provision watcher", http.StatusConflict)
		} else {
			LoggingClient.Error("Problem adding provision watcher: " + err.Error())
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return
	}

	// Notify Associates
	if err = notifyProvisionWatcherAssociates(pw, http.MethodPost); err != nil {
		LoggingClient.Error("Problem with notifying associating device services for the provision watcher: " + err.Error())
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(id))
}

// Update the provision watcher object
// ID is used first for identification, then name
// The service and profile cannot be updated
func restUpdateProvisionWatcher(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var from models.ProvisionWatcher
	if err := json.NewDecoder(r.Body).Decode(&from); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the provision watcher exists
	// Try by ID
	to, err := dbClient.GetProvisionWatcherById(from.Id)
	if err != nil {
		// Try by name
		if to, err = dbClient.GetProvisionWatcherByName(from.Name); err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Provision watcher not found", http.StatusNotFound)
				LoggingClient.Error("Provision watcher not found: " + err.Error())
			} else {
				LoggingClient.Error("Problem getting provision watcher: " + err.Error())
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			return
		}
	}

	if err := updateProvisionWatcherFields(from, &to, w); err != nil {
		LoggingClient.Error("Problem updating provision watcher: " + err.Error())
		return
	}

	if err := dbClient.UpdateProvisionWatcher(to); err != nil {
		LoggingClient.Error("Problem updating provision watcher: " + err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Notify Associates
	if err := notifyProvisionWatcherAssociates(to, http.MethodPut); err != nil {
		LoggingClient.Error("Problem notifying associated device services for provision watcher: " + err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the relevant fields of the provision watcher
func updateProvisionWatcherFields(from models.ProvisionWatcher, to *models.ProvisionWatcher, w http.ResponseWriter) error {
	if from.Identifiers != nil {
		to.Identifiers = from.Identifiers
	}
	if from.Origin != 0 {
		to.Origin = from.Origin
	}
	if from.Name != "" {
		// Check that the name is unique
		checkPW, err := dbClient.GetProvisionWatcherByName(from.Name)
		if err != nil {
			if err != db.ErrNotFound {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return err
			}
		}
		// Found one, compare the IDs to see if its another provision watcher
		if err != db.ErrNotFound {
			if checkPW.Id != to.Id {
				err = errors.New("Duplicate name for the provision watcher")
				http.Error(w, err.Error(), http.StatusConflict)
				return err
			}
		}
		to.Name = from.Name
	}

	return nil
}

// Notify the associated device services for the provision watcher
func notifyProvisionWatcherAssociates(pw models.ProvisionWatcher, action string) error {
	// Get the device service for the provision watcher
	ds, err := dbClient.GetDeviceServiceById(pw.Service.Service.Id)
	if err != nil {
		return err
	}

	// Notify the service
	if err = notifyAssociates([]models.DeviceService{ds}, pw.Id, action, models.PROVISIONWATCHER); err != nil {
		return err
	}

	return nil
}
