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
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"

	errors2 "github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device_profile"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

func restGetAllDeviceProfiles(w http.ResponseWriter, _ *http.Request) {
	res, err := dbClient.GetAllDeviceProfiles()
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(res) > Configuration.Service.MaxResultCount {
		err = errors.New("Max limit exceeded with request for profiles")
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		LoggingClient.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&res)
}

func restAddDeviceProfile(w http.ResponseWriter, r *http.Request) {
	var dp models.DeviceProfile

	if err := json.NewDecoder(r.Body).Decode(&dp); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if there are duplicate names in the device profile command list
	for _, c1 := range dp.CoreCommands {
		count := 0
		for _, c2 := range dp.CoreCommands {
			if c1.Name == c2.Name {
				count += 1
			}
		}
		if count > 1 {
			err := errors.New("Error adding device profile: Duplicate names in the commands")
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
	}

	id, err := dbClient.AddDeviceProfile(dp)
	if err != nil {
		if err == db.ErrNotUnique {
			http.Error(w, "Duplicate name for device profile", http.StatusConflict)
		} else if err == db.ErrNameEmpty {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(id))
}

func restUpdateDeviceProfile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var from models.DeviceProfile
	if err := json.NewDecoder(r.Body).Decode(&from); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	op := device_profile.NewUpdateDeviceProfileExecutor(dbClient, from)
	dp, err := op.Execute()
	if err != nil {
		LoggingClient.Error(err.Error())
		switch err.(type) {
		case errors2.ErrDeviceProfileNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case *errors2.ErrDuplicateName:
			http.Error(w, err.Error(), http.StatusConflict)
		case errors2.ErrDeviceProfileInvalidState:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	// Notify Associates
	err = notifyProfileAssociates(dp, dbClient, http.MethodPut)
	if err != nil {
		// Log the error but do not change the response to the client. We do not want this to affect the overall status
		// of the operation
		LoggingClient.Warn("Error while notifying profile associates of update: ", err.Error())
	}

	w.WriteHeader(http.StatusNoContent)
}

func restGetProfileByProfileId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars["id"]

	res, err := dbClient.GetDeviceProfileById(did)
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

func restDeleteProfileByProfileId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars["id"]

	// Check if the device profile exists
	dp, err := dbClient.GetDeviceProfileById(did)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		LoggingClient.Error(err.Error())
		return
	}

	// Delete the device profile
	if err = deleteDeviceProfile(dp, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Delete the device profile based on its name
func restDeleteProfileByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the device profile exists
	dp, err := dbClient.GetDeviceProfileByName(n)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	// Delete the device profile
	if err = deleteDeviceProfile(dp, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Delete the device profile
// Make sure there are no devices still using it
// Delete the associated commands
func deleteDeviceProfile(dp models.DeviceProfile, w http.ResponseWriter) error {
	// Check if the device profile is still in use by devices
	d, err := dbClient.GetDevicesByProfileId(dp.Id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}
	if len(d) > 0 {
		err = errors.New("Can't delete device profile, the profile is still in use by a device")
		http.Error(w, err.Error(), http.StatusConflict)
		return err
	}

	// Check if the device profile is still in use by provision watchers
	pw, err := dbClient.GetProvisionWatchersByProfileId(dp.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}
	if len(pw) > 0 {
		err = errors.New("Cant delete device profile, the profile is still in use by a provision watcher")
		http.Error(w, err.Error(), http.StatusConflict)
		return err
	}
	// Delete the profile
	if err := dbClient.DeleteDeviceProfileById(dp.Id); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	return nil
}

func restAddProfileByYaml(w http.ResponseWriter, r *http.Request) {
	f, _, err := r.FormFile("file")
	switch err {
	case nil:
	// do nothing
	case http.ErrMissingFile:
		err := errors.New("YAML file is empty")
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
		return
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
		return
	}
	if len(data) == 0 {
		err := errors.New("YAML file is empty")
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	addDeviceProfileYaml(data, w)
}

// Add a device profile with YAML content
// The YAML content is passed as a string in the http request
func restAddProfileByYamlRaw(w http.ResponseWriter, r *http.Request) {
	// Get the YAML string
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	addDeviceProfileYaml(body, w)
}

func addDeviceProfileYaml(data []byte, w http.ResponseWriter) {
	var dp models.DeviceProfile

	err := yaml.Unmarshal(data, &dp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		LoggingClient.Error(err.Error())
		return
	}

	// Check if there are duplicate names in the device profile command list
	for _, c1 := range dp.CoreCommands {
		count := 0
		for _, c2 := range dp.CoreCommands {
			if c1.Name == c2.Name {
				count += 1
			}
		}
		if count > 1 {
			err := errors.New("Error adding device profile: Duplicate names in the commands")
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
	}

	id, err := dbClient.AddDeviceProfile(dp)
	if err != nil {
		if err == db.ErrNotUnique {
			http.Error(w, "Duplicate profile name", http.StatusConflict)
		} else if err == db.ErrNameEmpty {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		LoggingClient.Error(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(id))
}

func restGetProfileByModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	an, err := url.QueryUnescape(vars[MODEL])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := dbClient.GetDeviceProfilesByModel(an)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetProfileWithLabel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	label, err := url.QueryUnescape(vars[LABEL])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := dbClient.GetDeviceProfilesWithLabel(label)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetProfileByManufacturerModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	man, err := url.QueryUnescape(vars[MANUFACTURER])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mod, err := url.QueryUnescape(vars[MODEL])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := dbClient.GetDeviceProfilesByManufacturerModel(man, mod)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetProfileByManufacturer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	man, err := url.QueryUnescape(vars[MANUFACTURER])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := dbClient.GetDeviceProfilesByManufacturer(man)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetProfileByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dn, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	// Get the device
	res, err := dbClient.GetDeviceProfileByName(dn)
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

func restGetYamlProfileByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	// Check for the device profile
	dp, err := dbClient.GetDeviceProfileByName(name)
	if err != nil {
		// Not found, return nil
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	// Marshal into yaml
	out, err := yaml.Marshal(dp)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

/*
 * Implementation: https://groups.google.com/forum/#!topic/golang-nuts/EZHtFOXA8UE
 * Response:
 * 	- 200: database generated identifier for the new device profile
 *	- 400: YAML file is empty
 *	- 409: an associated command's name is a duplicate for the profile or if the name is determined to not be uniqe with regard to others
 * 	- 503: Server Error
 */
func restGetYamlProfileById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars[ID]

	// Check if the device profile exists
	dp, err := dbClient.GetDeviceProfileById(id)
	if err != nil {
		if err == db.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(nil))
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	// Marshal the device profile into YAML
	out, err := yaml.Marshal(dp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

// Notify the associated device services for changes in the device profile
func notifyProfileAssociates(dp models.DeviceProfile, dl device.DeviceLoader, action string) error {
	// Get the devices
	op := device.NewProfileIdExecutor(Configuration.Service, dl, LoggingClient, dp.Id)
	d, err := op.Execute()
	if err != nil {
		LoggingClient.Error(err.Error())
		return err
	}

	// Get the services for each device
	// Use map as a Set
	dsMap := map[string]models.DeviceService{}
	ds := []models.DeviceService{}
	for _, device := range d {
		// Only add if not there
		if _, ok := dsMap[device.Service.Id]; !ok {
			dsMap[device.Service.Id] = device.Service
			ds = append(ds, device.Service)
		}
	}

	if err := notifyAssociates(ds, dp.Id, action, models.PROFILE); err != nil {
		LoggingClient.Error(err.Error())
		return err
	}

	return nil
}
