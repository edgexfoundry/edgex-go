/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
		LoggingClient.Error(err.Error())
		return
	}

	ctx := r.Context()
	body, status := commandByDeviceID(did, cid, string(b), p, ctx)
	if status != http.StatusOK {
		w.WriteHeader(status)
	} else {
		if len(body) > 0 {
			w.Header().Set("Content-Type", "application/json")
		}
		w.Write([]byte(body))
	}
}

func restPutDeviceAdminState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	did := vars[ID]
	as := vars[ADMINSTATE]
	ctx := r.Context()
	status, err := putDeviceAdminState(did, as, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())
	}
	w.WriteHeader(status)
}

func restPutDeviceAdminStateByDeviceName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dn := vars[NAME]
	as := vars[ADMINSTATE]
	ctx := r.Context()
	status, err := putDeviceAdminStateByName(dn, as, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())
	}
	w.WriteHeader(status)
}

func restPutDeviceOpState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	did := vars[ID]
	os := vars[OPSTATE]
	ctx := r.Context()
	status, err := putDeviceOpState(did, os, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())
	}
	w.WriteHeader(status)
}

func restPutDeviceOpStateByDeviceName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dn := vars[NAME]
	os := vars[OPSTATE]
	ctx := r.Context()
	status, err := putDeviceOpStateByName(dn, os, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())
	}
	w.WriteHeader(status)
}

func restGetCommandsByDeviceID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	did := vars[ID]
	ctx := r.Context()
	status, device, err := getCommandsByDeviceID(did, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	} else if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&device)
}

func restGetCommandsByDeviceName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dn := vars[NAME]
	ctx := r.Context()
	status, devices, err := getCommandsByDeviceName(dn, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	} else if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&devices)
}

func restGetAllCommands(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status, devices, err := getCommands(ctx)
	if err != nil {
		LoggingClient.Error(err.Error())
		w.WriteHeader(status)
	} else if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}
