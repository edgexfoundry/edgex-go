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

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/core/command/interfaces"
)

func restGetDeviceCommandByCommandID(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) {
	issueDeviceCommand(w, r, false, loggingClient, dbClient, deviceClient)
}

func restPutDeviceCommandByCommandID(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) {
	issueDeviceCommand(w, r, true, loggingClient, dbClient, deviceClient)
}

func issueDeviceCommand(
	w http.ResponseWriter,
	r *http.Request,
	isPutCommand bool,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	did := vars[ID]
	cid := vars[COMMANDID]
	b, err := ioutil.ReadAll(r.Body)
	if b == nil && err != nil {
		loggingClient.Error(err.Error())
		return
	}

	ctx := r.Context()
	body, status := commandByDeviceID(
		did,
		cid,
		string(b),
		r.URL.RawQuery,
		isPutCommand,
		ctx,
		loggingClient,
		dbClient,
		deviceClient)
	if status != http.StatusOK {
		http.Error(w, body, status)
	} else {
		if len(body) > 0 {
			w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
		}
		w.Write([]byte(body))
	}
}

func restGetDeviceCommandByNames(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) {
	issueDeviceCommandByNames(w, r, false, loggingClient, dbClient, deviceClient)
}

func restPutDeviceCommandByNames(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) {
	issueDeviceCommandByNames(w, r, true, loggingClient, dbClient, deviceClient)
}

func issueDeviceCommandByNames(
	w http.ResponseWriter,
	r *http.Request,
	isPutCommand bool,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	dn := vars[NAME]
	cn := vars[COMMANDNAME]

	ctx := r.Context()

	b, err := ioutil.ReadAll(r.Body)
	if b == nil && err != nil {
		loggingClient.Error(err.Error())
		return
	}
	body, status := commandByNames(
		dn,
		cn,
		string(b),
		r.URL.RawQuery,
		isPutCommand,
		ctx,
		loggingClient,
		dbClient,
		deviceClient)

	if status != http.StatusOK {
		http.Error(w, body, status)
	} else {
		if len(body) > 0 {
			w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
		}
		w.Write([]byte(body))
	}
}

func restGetCommandsByDeviceID(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) {
	vars := mux.Vars(r)
	did := vars[ID]
	ctx := r.Context()
	status, device, err := getCommandsByDeviceID(did, ctx, loggingClient, dbClient, deviceClient)
	if err != nil {
		loggingClient.Error(err.Error())
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	} else if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(&device)
}

func restGetCommandsByDeviceName(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) {
	vars := mux.Vars(r)
	dn := vars[NAME]
	ctx := r.Context()
	status, devices, err := getCommandsByDeviceName(dn, ctx, loggingClient, dbClient, deviceClient)
	if err != nil {
		loggingClient.Error(err.Error())
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	} else if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(&devices)
}

func restGetAllCommands(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) {
	ctx := r.Context()
	status, devices, err := getCommands(ctx, loggingClient, dbClient, deviceClient)
	if err != nil {
		loggingClient.Error(err.Error())
		w.WriteHeader(status)
	} else if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(devices)
}
