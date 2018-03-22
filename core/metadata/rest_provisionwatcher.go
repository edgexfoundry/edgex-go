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
 * @microservice: core-metadata-go service
 * @author: Spencer Bull & Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package metadata

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
)

func restGetProvisionWatchers(w http.ResponseWriter, _ *http.Request) {
	res := make([]models.ProvisionWatcher, 0)
	if err := getAllProvisionWatchers(&res); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	// Check the length
	if len(res) > configuration.ReadMaxLimit {
		err := errors.New("Max limit exceeded")
		loggingClient.Error(err.Error(), "")
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
	var pw models.ProvisionWatcher
	if err := getProvisionWatcherById(&pw, id); err != nil {
		errMessage := "Provision Watcher not found by ID: " + err.Error()
		loggingClient.Error(errMessage, "")
		http.Error(w, errMessage, http.StatusNotFound)
		return
	}

	if err := deleteProvisionWatcher(pw, w); err != nil {
		loggingClient.Error("Error deleting provision watcher", "")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("true"))
}

func restDeleteProvisionWatcherByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	// Check if the provision watcher exists
	var pw models.ProvisionWatcher
	if err = getProvisionWatcherByName(&pw, n); err != nil {
		if err == mgo.ErrNotFound {
			errMessage := "Provision watcher not found: " + err.Error()
			http.Error(w, errMessage, http.StatusNotFound)
			loggingClient.Error(errMessage, "")
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	if err = deleteProvisionWatcher(pw, w); err != nil {
		loggingClient.Error("Problem deleting provision watcher: "+err.Error(), "")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Delete the provision watcher
func deleteProvisionWatcher(pw models.ProvisionWatcher, w http.ResponseWriter) error {
	if err := deleteById(PWCOL, pw.Id.Hex()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	if err := notifyProvisionWatcherAssociates(pw, http.MethodDelete); err != nil {
		loggingClient.Error("Problem notifying associated device services to provision watcher: "+err.Error(), "")
	}

	return nil
}

func restGetProvisionWatcherById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var id string = vars[ID]
	var res models.ProvisionWatcher

	if err := getProvisionWatcherById(&res, id); err != nil {
		if err == mgo.ErrNotFound {
			errMessage := "Problem getting provision watcher by ID: " + err.Error()
			loggingClient.Error(errMessage, "")
			http.Error(w, errMessage, http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error(), "")
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
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}
	var res models.ProvisionWatcher

	err = getProvisionWatcherByName(&res, n)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			loggingClient.Error("Provision watcher not found: "+err.Error(), "")
		} else {
			loggingClient.Error(err.Error(), "")
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
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
	var dp models.DeviceProfile
	if err := getDeviceProfileById(&dp, pid); err != nil {
		loggingClient.Error("Device profile not found: "+err.Error(), "")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	res := make([]models.ProvisionWatcher, 0)
	err := getProvisionWatcherByProfileId(&res, pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Problem getting provision watcher: "+err.Error(), "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
func restGetProvisionWatchersByProfileName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pn, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	// Check if the device profile exists
	var dp models.DeviceProfile
	if err = getDeviceProfileByName(&dp, pn); err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, "Device profile not found", http.StatusNotFound)
			loggingClient.Error("Device profile not found: "+err.Error(), "")
		} else {
			loggingClient.Error("Problem getting device profile: "+err.Error(), "")
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return
	}

	res := make([]models.ProvisionWatcher, 0)
	err = getProvisionWatcherByProfileId(&res, dp.Id.Hex())
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Problem getting provision watcher: "+err.Error(), "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
func restGetProvisionWatchersByServiceId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var sid string = vars[ID]

	// Check if the device service exists
	var ds models.DeviceService
	if err := getDeviceServiceById(&ds, sid); err != nil {
		http.Error(w, "Device Service not found", http.StatusNotFound)
		loggingClient.Error("Device service not found: "+err.Error(), "")
		return
	}

	res := make([]models.ProvisionWatcher, 0)
	err := getProvisionWatchersByServiceId(&res, sid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Problem getting provision watcher: "+err.Error(), "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
func restGetProvisionWatchersByServiceName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sn, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	// Check if the device service exists
	var ds models.DeviceService
	if err = getDeviceServiceByName(&ds, sn); err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, "Device service not found", http.StatusNotFound)
			loggingClient.Error("Device service not found: "+err.Error(), "")
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error("Problem getting device service: "+err.Error(), "")
		}
		return
	}

	// Get the provision watchers
	res := make([]models.ProvisionWatcher, 0)
	err = getProvisionWatchersByServiceId(&res, ds.Service.Id.Hex())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		loggingClient.Error("Problem getting provision watcher: "+err.Error(), "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
func restGetProvisionWatchersByIdentifier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	k, err := url.QueryUnescape(vars[KEY])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}
	v, err := url.QueryUnescape(vars[VALUE])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	res := make([]models.ProvisionWatcher, 0)
	if err := getProvisionWatchersByIdentifier(&res, k, v); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Problem getting provision watchers: "+err.Error(), "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
func restAddProvisionWatcher(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var pw models.ProvisionWatcher
	if err := json.NewDecoder(r.Body).Decode(&pw); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the name exists
	if pw.Name == "" {
		err := errors.New("No name provided for new provision watcher")
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// Check if the device profile exists
	// Try by ID
	if err := getDeviceProfileById(&pw.Profile, pw.Profile.Id.Hex()); err != nil {
		// Try by name
		if err = getDeviceProfileByName(&pw.Profile, pw.Profile.Name); err != nil {
			if err == mgo.ErrNotFound {
				loggingClient.Error("Device profile not found for provision watcher: "+err.Error(), "")
				http.Error(w, "Device profile not found for provision watcher", http.StatusConflict)
			} else {
				loggingClient.Error("Problem getting device profile: "+err.Error(), "")
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			return
		}
	}

	// Check if the device service exists
	// Try by ID
	if err := getDeviceServiceById(&pw.Service, pw.Service.Service.Id.Hex()); err != nil {
		// Try by name
		if err = getDeviceServiceByName(&pw.Service, pw.Service.Service.Name); err != nil {
			if err == mgo.ErrNotFound {
				http.Error(w, "Device service not found for provision watcher", http.StatusConflict)
				loggingClient.Error("Device service not found for provision watcher: "+err.Error(), "")
			} else {
				loggingClient.Error("Problem getting device service for provision wathcer: "+err.Error(), "")
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			return
		}
	}

	if err := addProvisionWatcher(&pw); err != nil {
		if err == ErrDuplicateName {
			loggingClient.Error("Duplicate name for the provision watcher: "+err.Error(), "")
			http.Error(w, "Duplicate name for the provision watcher", http.StatusConflict)
		} else {
			loggingClient.Error("Problem adding provision watcher: "+err.Error(), "")
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return
	}

	// Notify Associates
	if err := notifyProvisionWatcherAssociates(pw, http.MethodPost); err != nil {
		loggingClient.Error("Problem with notifying associating device services for the provision watcher: "+err.Error(), "")
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pw.Id.Hex()))
}

// Update the provision watcher object
// ID is used first for identification, then name
// The service and profile cannot be updated
func restUpdateProvisionWatcher(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var from models.ProvisionWatcher
	if err := json.NewDecoder(r.Body).Decode(&from); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the provision watcher exists
	var to models.ProvisionWatcher
	// Try by ID
	if err := getProvisionWatcherById(&to, from.Id.Hex()); err != nil {
		// Try by name
		if err = getProvisionWatcherByName(&to, from.Name); err != nil {
			if err == mgo.ErrNotFound {
				http.Error(w, "Provision watcher not found", http.StatusNotFound)
				loggingClient.Error("Provision watcher not found: "+err.Error(), "")
			} else {
				loggingClient.Error("Problem getting provision watcher: "+err.Error(), "")
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			return
		}
	}

	if err := updateProvisionWatcherFields(from, &to, w); err != nil {
		loggingClient.Error("Problem updating provision watcher: "+err.Error(), "")
		return
	}

	if err := updateProvisionWatcher(to); err != nil {
		loggingClient.Error("Problem updating provision watcher: "+err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Notify Associates
	if err := notifyProvisionWatcherAssociates(to, http.MethodPut); err != nil {
		loggingClient.Error("Problem notifying associated device services for provision watcher: "+err.Error(), "")
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
		var checkPW models.ProvisionWatcher
		err := getProvisionWatcherByName(&checkPW, from.Name)
		if err != nil {
			if err != mgo.ErrNotFound {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return err
			}
		}
		// Found one, compare the IDs to see if its another provision watcher
		if err != mgo.ErrNotFound {
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
	var ds models.DeviceService
	if err := getDeviceServiceById(&ds, pw.Service.Service.Id.Hex()); err != nil {
		return err
	}

	var services []models.DeviceService
	services = append(services, ds)

	// Notify the service
	if err := notifyAssociates(services, pw.Id.Hex(), action, models.PROVISIONWATCHER); err != nil {
		return err
	}

	return nil
}
