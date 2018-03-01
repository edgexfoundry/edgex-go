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

func restGetAllCommands(w http.ResponseWriter, _ *http.Request) {
	results := make([]models.Command, 0)
	err := getAllCommands(&results)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(results) > configuration.ReadMaxLimit {
		loggingClient.Error(err.Error(), "")
		http.Error(w, errors.New("Max limit exceeded").Error(), http.StatusRequestEntityTooLarge)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&results)
}

func restAddCommand(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var c models.Command
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	if err := addCommand(&c); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(c.Id.Hex()))
}

// Update a command
// 404 if the command can't be found by ID
// 409 if the name of the command changes and its not unique to the device profile
func restUpdateCommand(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var c models.Command
	var res models.Command
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if command exists (By ID)
	err := getCommandById(&res, c.Id.Hex())
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Name is changed, make sure the new name doesn't conflict with device profile
	if c.Name != "" {
		var dp []models.DeviceProfile
		err = getDeviceProfilesUsingCommand(&dp, c)
		if err != nil {
			loggingClient.Error(err.Error(), "")
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		// Loop through matched device profiles to ensure the name isn't duplicate
		for _, profile := range dp {
			for _, command := range profile.Commands {
				if command.Name == c.Name && command.Id != c.Id {
					err = errors.New("Error updating command: duplicate command name in device profile")
					loggingClient.Error(err.Error(), "")
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
			}
		}
	}

	if err := updateCommand(&c, &res); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restGetCommandById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars[ID]
	var res models.Command
	err := getCommandById(&res, did)
	if err != nil {
		if err == ErrNotFound {
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

func restGetCommandByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	results := []models.Command{}
	err = getCommandByName(&results, n)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// Delete a command by its ID
func restDeleteCommandById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var id string = vars[ID]

	// Check if the command exists
	var c models.Command
	err := getCommandById(&c, id)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Check if the command is still in use by a device profile
	isStillInUse, err := isCommandStillInUse(c)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	if isStillInUse {
		err = errors.New("Can't delete command.  Its still in use by Device Profiles")
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	if err := deleteCommandById(id); err != nil {
		loggingClient.Error(err.Error(), "")
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if err == ErrCommandStillInUse {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Helper function to determine if the command is still in use by device profiles
func isCommandStillInUse(c models.Command) (bool, error) {
	var dp []models.DeviceProfile
	err := getDeviceProfilesUsingCommand(&dp, c)
	if err != nil {
		return false, err
	}
	if len(dp) == 0 {
		return false, err
	}

	return true, err
}
