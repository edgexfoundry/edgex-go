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
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/yaml.v2"
)

func restGetAllDeviceProfiles(w http.ResponseWriter, _ *http.Request) {
	res := []models.DeviceProfile{}
	if err := getAllDeviceProfiles(&res); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(res) > configuration.ReadMaxLimit {
		err := errors.New("Max limit exceeded with request for profiles")
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		loggingClient.Error(err.Error(), "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&res)
}

func restAddDeviceProfile(w http.ResponseWriter, r *http.Request) {
	var dp models.DeviceProfile

	if err := json.NewDecoder(r.Body).Decode(&dp); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if there are duplicate names in the device profile command list
	for _, c1 := range dp.Commands {
		count := 0
		for _, c2 := range dp.Commands {
			if c1.Name == c2.Name {
				count += 1
			}
		}
		if count > 1 {
			err := errors.New("Error adding device profile: Duplicate names in the commands")
			loggingClient.Error(err.Error(), "")
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
	}

	if err := addDeviceProfile(&dp); err != nil {
		if err == ErrDuplicateName {
			http.Error(w, "Duplicate name for device profile", http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(dp.Id.Hex()))
}

func restUpdateDeviceProfile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var from models.DeviceProfile
	if err := json.NewDecoder(r.Body).Decode(&from); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the Device Profile exists
	var to models.DeviceProfile
	// First try with ID
	err := getDeviceProfileById(&to, from.Id.Hex())
	if err != nil {
		// Try with name
		err = getDeviceProfileByName(&to, from.Name)
		if err != nil {
			loggingClient.Error(err.Error(), "")
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}

	// Update the device profile fields based on the passed JSON
	if err := updateDeviceProfileFields(from, &to, w); err != nil {
		loggingClient.Error(err.Error(), "")
		return
	}
	if err := updateDeviceProfile(&to); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the fields of the device profile
// to - the device profile that was already in Mongo (whose fields we're updating)
// from - the device profile that was passed in with the request
func updateDeviceProfileFields(from models.DeviceProfile, to *models.DeviceProfile, w http.ResponseWriter) error {
	if from.Description != "" {
		to.Description = from.Description
	}
	if from.Labels != nil {
		to.Labels = from.Labels
	}
	if from.Manufacturer != "" {
		to.Manufacturer = from.Manufacturer
	}
	if from.Model != "" {
		to.Model = from.Model
	}
	if from.Objects != nil {
		to.Objects = from.Objects
	}
	if from.Origin != 0 {
		to.Origin = from.Origin
	}
	if from.Name != "" {
		to.Name = from.Name
		// Names must be unique for each device profile
		if err := checkDuplicateProfileNames(*to, w); err != nil {
			return err
		}
	}
	if from.DeviceResources != nil {
		to.DeviceResources = from.DeviceResources
	}
	if from.Resources != nil {
		to.Resources = from.Resources
	}
	if from.Commands != nil {
		// Check for duplicates by command name
		if err := checkDuplicateCommands(from, w); err != nil {
			return err
		}

		// taking lazy approach to commands - remove them all and add them
		// all back in. TODO - someday make this a two phase commit so
		// commands don't get wiped out before profile

		// Delete the old commands
		if err := deleteCommands(*to, w); err != nil {
			return err
		}

		to.Commands = from.Commands

		// Add the new commands
		if err := addCommands(to, w); err != nil {
			return err
		}
	}

	return nil
}

// Check for duplicate names in device profiles
func checkDuplicateProfileNames(dp models.DeviceProfile, w http.ResponseWriter) error {
	profiles := []models.DeviceProfile{}
	if err := getAllDeviceProfiles(&profiles); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	for _, p := range profiles {
		if p.Name == dp.Name && p.Id != dp.Id {
			err := errors.New("Duplicate profile name")
			http.Error(w, err.Error(), http.StatusConflict)
			return err
		}
	}

	return nil
}

// Check for duplicate command names in the device profile
func checkDuplicateCommands(dp models.DeviceProfile, w http.ResponseWriter) error {
	// Check if there are duplicate names in the device profile command list
	for _, c1 := range dp.Commands {
		count := 0
		for _, c2 := range dp.Commands {
			if c1.Name == c2.Name {
				count += 1
			}
		}
		if count > 1 {
			err := errors.New("Error adding device profile: Duplicate names in the commands")
			http.Error(w, err.Error(), http.StatusConflict)
			return err
		}
	}

	return nil
}

// Delete all of the commands that are a part of the device profile
func deleteCommands(dp models.DeviceProfile, w http.ResponseWriter) error {
	for _, command := range dp.Commands {
		err := deleteCommandById(command.Id.Hex())
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return err
		}
	}

	return nil
}

// Add all of the commands that are a part of the device profile
func addCommands(dp *models.DeviceProfile, w http.ResponseWriter) error {
	for i := range dp.Commands {
		if err := addCommand(&(dp.Commands[i])); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return err
		}
	}

	return nil
}

func restGetProfileByProfileId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars["id"]
	var res models.DeviceProfile
	if err := getDeviceProfileById(&res, did); err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restDeleteProfileByProfileId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars["id"]

	// Check if the device profile exists
	var dp models.DeviceProfile
	if err := getDeviceProfileById(&dp, did); err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	// Delete the device profile
	if err := deleteDeviceProfile(dp, w); err != nil {
		loggingClient.Error(err.Error(), "")
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
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device profile exists
	var dp models.DeviceProfile
	if err = getDeviceProfileByName(&dp, n); err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	// Delete the device profile
	if err = deleteDeviceProfile(dp, w); err != nil {
		loggingClient.Error(err.Error(), "")
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
	var d []models.Device
	if err := getDevicesByProfileId(&d, dp.Id.Hex()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}
	if len(d) > 0 {
		err := errors.New("Can't delete device profile, the profile is still in use by a device")
		http.Error(w, err.Error(), http.StatusConflict)
		return err
	}

	// Check if the device profile is still in use by provision watchers
	var pw []models.ProvisionWatcher
	if err := getProvisionWatcherByProfileId(&pw, dp.Id.Hex()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}
	if len(pw) > 0 {
		err := errors.New("Cant delete device profile, the profile is still in use by a provision watcher")
		http.Error(w, err.Error(), http.StatusConflict)
		return err
	}

	// Delete the profile
	if err := deleteDeviceProfileById(dp.Id.Hex()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	// TODO: Notify Associates
	if err := notifyProfileAssociates(dp, http.MethodDelete); err != nil {
		loggingClient.Error(err.Error(), "")
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
		loggingClient.Error(err.Error(), "")
		return
	default:
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
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
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	addDeviceProfileYaml(body, w)
}

func addDeviceProfileYaml(data []byte, w http.ResponseWriter) {
	var dp models.DeviceProfile

	err := yaml.Unmarshal(data, &dp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	// Check if there are duplicate names in the device profile command list
	for _, c1 := range dp.Commands {
		count := 0
		for _, c2 := range dp.Commands {
			if c1.Name == c2.Name {
				count += 1
			}
		}
		if count > 1 {
			err := errors.New("Error adding device profile: Duplicate names in the commands")
			loggingClient.Error(err.Error(), "")
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
	}

	if err := addDeviceProfile(&dp); err != nil {
		if err == ErrDuplicateName {
			http.Error(w, "Duplicate profile name", http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(dp.Id.Hex()))
}

func restGetProfileByModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	an, err := url.QueryUnescape(vars[MODEL])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	res := make([]models.DeviceProfile, 0)
	if err := getDeviceProfilesByModel(&res, an); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetProfileWithLabel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	label, err := url.QueryUnescape(vars[LABEL])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	var labels []string
	labels = append(labels, label)
	res := make([]models.DeviceProfile, 0)
	if err := getDeviceProfilesWithLabel(&res, labels); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetProfileByManufacturerModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	man, err := url.QueryUnescape(vars[MANUFACTURER])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	mod, err := url.QueryUnescape(vars[MODEL])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	res := make([]models.DeviceProfile, 0)
	if err := getDeviceProfilesByManufacturerModel(&res, man, mod); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetProfileByManufacturer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	man, err := url.QueryUnescape(vars[MANUFACTURER])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	res := make([]models.DeviceProfile, 0)
	if err := getDeviceProfilesByManufacturer(&res, man); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetProfileByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dn, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	// Get the device
	var res models.DeviceProfile
	if err := getDeviceProfileByName(&res, dn); err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetYamlProfileByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	// Check for the device profile
	var dp models.DeviceProfile
	err = getDeviceProfileByName(&dp, name)
	if err != nil {
		// Not found, return nil
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	// Marshal into yaml
	out, err := yaml.Marshal(dp)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
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
	var dp models.DeviceProfile
	err := getDeviceProfileById(&dp, id)
	if err != nil {
		if err == mgo.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(nil))
			return
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	// Marshal the device profile into YAML
	out, err := yaml.Marshal(dp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

// Notify the associated device services for the addressable
func notifyProfileAssociates(dp models.DeviceProfile, action string) error {
	// Get the devices
	var d []models.Device
	if err := getDevicesByProfileId(&d, dp.Id.Hex()); err != nil {
		loggingClient.Error(err.Error(), "")
		return err
	}

	// Get the services for each device
	// Use map as a Set
	dsMap := map[string]models.DeviceService{}
	var ds []models.DeviceService
	for _, device := range d {
		// Only add if not there
		if _, ok := dsMap[device.Service.Service.Id.Hex()]; !ok {
			dsMap[device.Service.Service.Id.Hex()] = device.Service
			ds = append(ds, device.Service)
		}
	}

	if err := notifyAssociates(ds, dp.Id.Hex(), action, models.PROFILE); err != nil {
		loggingClient.Error(err.Error(), "")
		return err
	}

	return nil
}
