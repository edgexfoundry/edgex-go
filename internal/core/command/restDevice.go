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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/requests/states/admin"
	"github.com/edgexfoundry/go-mod-core-contracts/requests/states/operating"
	"github.com/gorilla/mux"
)

func restGetDeviceCommandByCommandID(w http.ResponseWriter, r *http.Request) {
	issueDeviceCommand(w, r, false)
}

func restPutDeviceCommandByCommandID(w http.ResponseWriter, r *http.Request) {
	issueDeviceCommand(w, r, true)
}

func issueDeviceCommand(w http.ResponseWriter, r *http.Request, isPutCommand bool) {
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
	body, status := commandByDeviceID(did, cid, string(b), isPutCommand, ctx)
	if status != http.StatusOK {
		http.Error(w, body, status)
	} else {
		if len(body) > 0 {
			w.Header().Set("Content-Type", "application/json")
		}
		w.Write([]byte(body))
	}
}

func restGetDeviceCommandByNames(w http.ResponseWriter, r *http.Request) {
	issueDeviceCommandByNames(w, r, false)
}

func restPutDeviceCommandByNames(w http.ResponseWriter, r *http.Request) {
	issueDeviceCommandByNames(w, r, true)
}

func issueDeviceCommandByNames(w http.ResponseWriter, r *http.Request, isPutCommand bool) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	dn := vars[NAME]
	cn := vars[COMMANDNAME]

	ctx := r.Context()

	b, err := ioutil.ReadAll(r.Body)
	if b == nil && err != nil {
		LoggingClient.Error(err.Error())
		return
	}
	body, status := commandByNames(dn, cn, string(b), isPutCommand, ctx)

	if status != http.StatusOK {
		http.Error(w, body, status)
	} else {
		if len(body) > 0 {
			w.Header().Set("Content-Type", "application/json")
		}
		w.Write([]byte(body))
	}
}

func decodeState(r *http.Request) (mode string, state string, err error) {
	var admin admin.UpdateRequest
	var ops operating.UpdateRequest

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	decoder := json.NewDecoder(bytes.NewBuffer(bodyBytes))
	err = decoder.Decode(&admin)
	if err != nil {
		LoggingClient.Error(err.Error())
		return
	}

	if admin.AdminState != "" {
		return ADMINSTATE, string(admin.AdminState), nil
	}

	// In this case, the supplied request was not for the AdminState. Try OperatingState.
	decoder = json.NewDecoder(bytes.NewBuffer(bodyBytes))
	err = decoder.Decode(&ops)
	if err != nil {
		LoggingClient.Error(err.Error())
		return
	}
	if ops.OperatingState != "" {
		return OPSTATE, string(ops.OperatingState), nil
	}

	// In this case, the request we were given in completely invalid
	return "","", fmt.Errorf("unknown request type")
}

func restPutDeviceStateByDeviceId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	did := vars[ID]

	updateMode, state, err := decodeState(r)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	var status int

	switch updateMode {
	case ADMINSTATE:
		status, err = putDeviceAdminState(did, state, ctx)
		break
	case OPSTATE:
		status, err = putDeviceOpState(did, state, ctx)
		break
	}
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
}

func restPutDeviceStateByDeviceName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dn := vars[NAME]
	ctx := r.Context()

	updateMode, state, err := decodeState(r)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var status int

	switch updateMode {
	case ADMINSTATE:
		status, err = putDeviceAdminStateByName(dn, state, ctx)
		break
	case OPSTATE:
		status, err = putDeviceOpStateByName(dn, state, ctx)
		break
	}
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
