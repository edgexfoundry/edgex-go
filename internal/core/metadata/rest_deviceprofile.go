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
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"

	errors2 "github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device_profile"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

func restGetAllDeviceProfiles(w http.ResponseWriter, _ *http.Request) {
	op := device_profile.NewGetAllExecutor(Configuration.Service, dbClient, LoggingClient)
	res, err := op.Execute()
	if err != nil {
		LoggingClient.Error(err.Error())
		switch err.(type) {
		case errors.ErrLimitExceeded:
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

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
			err := errors.NewErrDuplicateName("Error adding device profile: Duplicate names in the commands")
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
	}

	if Configuration.Writable.EnableValueDescriptorManagement {
		op := device_profile.NewAddExecutor(r.Context(), vdc, LoggingClient, dp.DeviceResources...)
		err := op.Execute()
		if err != nil {
			LoggingClient.Error(err.Error())
			switch err.(type) {
			case types.ErrServiceClient:
				http.Error(w, err.Error(), err.(types.ErrServiceClient).StatusCode)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

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

	if Configuration.Writable.EnableValueDescriptorManagement {
		vdOp := device_profile.NewUpdateValueDescriptorExecutor(from, dbClient, vdc, LoggingClient, r.Context())
		err := vdOp.Execute()
		if err != nil {
			LoggingClient.Error(err.Error())
			switch err.(type) {
			case types.ErrServiceClient:
				http.Error(w, err.Error(), err.(types.ErrServiceClient).StatusCode)
			case errors.ErrDeviceProfileNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
			case *errors2.ErrValueDescriptorsInUse:
				http.Error(w, err.Error(), http.StatusConflict)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			return
		}
	}

	op := device_profile.NewUpdateDeviceProfileExecutor(dbClient, from)
	dp, err := op.Execute()
	if err != nil {
		LoggingClient.Error(err.Error())
		switch err.(type) {
		case errors.ErrDeviceProfileNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.ErrDeviceProfileInvalidState:
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restGetProfileByProfileId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did = vars["id"]

	op := device_profile.NewGetProfileID(did, dbClient)
	res, err := op.Execute()
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
	var did = vars["id"]

	op := device_profile.NewDeleteByIDExecutor(dbClient, did)
	err := op.Execute()
	if err != nil {
		LoggingClient.Error(err.Error())
		switch err.(type) {
		case errors.ErrDeviceProfileNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.ErrDeviceProfileInvalidState:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

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

	op := device_profile.NewDeleteByNameExecutor(dbClient, n)
	err = op.Execute()
	if err != nil {
		LoggingClient.Error(err.Error())
		switch err.(type) {
		case errors.ErrDeviceProfileNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.ErrDeviceProfileInvalidState:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restAddProfileByYaml(w http.ResponseWriter, r *http.Request) {
	f, _, err := r.FormFile("file")
	if err != nil {
		switch err {
		case http.ErrMissingFile:
			err := errors.NewErrEmptyFile("YAML")
			http.Error(w, err.Error(), http.StatusBadRequest)
			LoggingClient.Error(err.Error())
			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
			return
		}
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
		return
	}
	if len(data) == 0 {
		err := errors.NewErrEmptyFile("YAML")
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	var dp models.DeviceProfile

	err = yaml.Unmarshal(data, &dp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		LoggingClient.Error(err.Error())
		return
	}

	op := device_profile.NewAddDeviceProfileExecutor(dp, dbClient)
	id, err := op.Execute()

	if err != nil {
		switch err.(type) {
		case models.ErrContractInvalid:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.ErrDeviceProfileInvalidState:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case *errors.ErrDuplicateName:
			http.Error(w, err.Error(), http.StatusConflict)
		case errors.ErrEmptyDeviceProfileName:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(id))
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

	var dp models.DeviceProfile

	err = yaml.Unmarshal(body, &dp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		LoggingClient.Error(err.Error())
		return
	}

	op := device_profile.NewAddDeviceProfileExecutor(dp, dbClient)
	id, err := op.Execute()

	if err != nil {
		switch err.(type) {
		case models.ErrContractInvalid:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.ErrDeviceProfileInvalidState:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case *errors.ErrDuplicateName:
			http.Error(w, err.Error(), http.StatusConflict)
		case errors.ErrEmptyDeviceProfileName:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

	op := device_profile.NewGetModelExecutor(an, dbClient)
	res, err := op.Execute()
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

	op := device_profile.NewGetLabelExecutor(label, dbClient)
	res, err := op.Execute()
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

	op := device_profile.NewGetManufacturerModelExecutor(man, mod, dbClient)
	res, err := op.Execute()
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

	op := device_profile.NewGetManufacturerExecutor(man, dbClient)
	res, err := op.Execute()
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
	op := device_profile.NewGetProfileName(dn, dbClient)
	res, err := op.Execute()
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
	op := device_profile.NewGetProfileName(name, dbClient)
	dp, err := op.Execute()
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
	op := device_profile.NewGetProfileID(id, dbClient)
	dp, err := op.Execute()
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
