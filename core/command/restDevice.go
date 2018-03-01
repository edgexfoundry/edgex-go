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
 * @microservice: core-command-go service
 * @author: Spencer Bull, Dell
 * @version: 0.5.0
 *******************************************************************************/
package command

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	mux "github.com/gorilla/mux"
)

func restGetDeviceCommandByCommandID(w http.ResponseWriter, r *http.Request) {
	issueDeviceCommand(w, r, false)
}

func restPutDeviceCommandByCommandID(w http.ResponseWriter, r *http.Request) {
	issueDeviceCommand(w, r, true)
}

func issueDeviceCommand(w http.ResponseWriter, r *http.Request, p bool) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	did := vars[ID]
	cid := vars[COMMANDID]
	b, err := ioutil.ReadAll(r.Body)
	if b == nil && err != nil {
		loggingClient.Error(err.Error(), "")
		return
	}
	body, status := commandByDeviceID(did, cid, string(b), p)
	if status != 200 {
		w.WriteHeader(status)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}
}

func restPutDeviceAdminState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	did := vars[ID]
	as := vars[ADMINSTATE]
	status, err := putDeviceAdminState(did, as)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	w.WriteHeader(status)

}

func restPutDeviceAdminStateByDeviceName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dn := vars[NAME]
	as := vars[ADMINSTATE]
	status, err := putDeviceAdminStateByName(dn, as)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	w.WriteHeader(status)
}

func restPutDeviceOpState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	did := vars[ID]
	os := vars[OPSTATE]
	status, err := putDeviceOpState(did, os)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	w.WriteHeader(status)
}

func restPutDeviceOpStateByDeviceName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dn := vars[NAME]
	os := vars[OPSTATE]
	status, err := putDeviceOpStateByName(dn, os)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	w.WriteHeader(status)
}

func restGetCommandsByDeviceID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	did := vars[ID]
	status, device, err := getCommandsByDeviceID(did)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	} else if status != http.StatusOK {
		w.WriteHeader(status)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&device)
}

func restGetCommandsByDeviceName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dn := vars[NAME]
	status, devices, err := getCommandsByDeviceName(dn)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	} else if status != http.StatusOK {
		w.WriteHeader(status)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&devices)
}

func restGetAllCommands(w http.ResponseWriter, _ *http.Request) {
	status, devices, err := getCommands()
	if err != nil {
		loggingClient.Error(err.Error(), "")
		w.WriteHeader(http.StatusServiceUnavailable)
	} else if status != http.StatusOK {
		w.WriteHeader(status)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}
