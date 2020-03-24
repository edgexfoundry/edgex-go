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

	"github.com/edgexfoundry/edgex-go/internal/core/command/config"
	"github.com/edgexfoundry/edgex-go/internal/core/command/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"
)

func restGetDeviceCommandByCommandID(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	httpErrorHandler errorconcept.ErrorHandler) {

	issueDeviceCommand(w, r, false, lc, dbClient, deviceClient, httpErrorHandler)
}

func restPutDeviceCommandByCommandID(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	httpErrorHandler errorconcept.ErrorHandler) {

	issueDeviceCommand(w, r, true, lc, dbClient, deviceClient, httpErrorHandler)
}

func issueDeviceCommand(
	w http.ResponseWriter,
	r *http.Request,
	isPutCommand bool,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	httpErrorHandler errorconcept.ErrorHandler) {

	defer r.Body.Close()

	vars := mux.Vars(r)
	did := vars[ID]
	cid := vars[COMMANDID]
	b, err := ioutil.ReadAll(r.Body)
	if b == nil && err != nil {
		lc.Error(err.Error())
		return
	}

	ctx := r.Context()
	body, err := executeCommandByDeviceID(
		ctx,
		did,
		cid,
		string(b),
		r.URL.RawQuery,
		isPutCommand,
		lc,
		dbClient,
		deviceClient)
	if err != nil {
		httpErrorHandler.HandleManyVariants(
			w,
			err,
			[]errorconcept.ErrorConceptType{
				errorconcept.NewServiceClientHttpError(err),
				errorconcept.Device.Locked,
				errorconcept.Database.NotFound,
				errorconcept.Command.NotAssociatedWithDevice,
			},
			errorconcept.Default.InternalServerError)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(body))
}

func restGetDeviceCommandByNames(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	httpErrorHandler errorconcept.ErrorHandler) {

	issueDeviceCommandByNames(w, r, false, lc, dbClient, deviceClient, httpErrorHandler)
}

func restPutDeviceCommandByNames(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	httpErrorHandler errorconcept.ErrorHandler) {

	issueDeviceCommandByNames(w, r, true, lc, dbClient, deviceClient, httpErrorHandler)
}

func issueDeviceCommandByNames(
	w http.ResponseWriter,
	r *http.Request,
	isPutCommand bool,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	httpErrorHandler errorconcept.ErrorHandler) {

	defer r.Body.Close()

	vars := mux.Vars(r)
	dn := vars[NAME]
	cn := vars[COMMANDNAME]

	ctx := r.Context()

	b, err := ioutil.ReadAll(r.Body)
	if b == nil && err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}
	body, err := executeCommandByName(
		ctx,
		dn,
		cn,
		string(b),
		r.URL.RawQuery,
		isPutCommand,
		lc,
		dbClient,
		deviceClient)

	if err != nil {
		httpErrorHandler.HandleManyVariants(
			w,
			err,
			[]errorconcept.ErrorConceptType{
				errorconcept.NewServiceClientHttpError(err),
				errorconcept.Device.Locked,
				errorconcept.Database.NotFound,
			},
			errorconcept.Default.InternalServerError)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(body))
}

func restGetCommandsByDeviceID(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	configuration *config.ConfigurationStruct,
	httpErrorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	did := vars[ID]
	ctx := r.Context()
	device, err := getCommandsByDeviceID(ctx, did, dbClient, deviceClient, configuration)
	if err != nil {
		httpErrorHandler.HandleManyVariants(
			w,
			err,
			[]errorconcept.ErrorConceptType{
				errorconcept.NewServiceClientHttpError(err),
				errorconcept.Device.NotFoundInDB,
				errorconcept.Database.NotFound,
			},
			errorconcept.Default.InternalServerError)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(&device)
}

func restGetCommandsByDeviceName(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	configuration *config.ConfigurationStruct,
	httpErrorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	dn := vars[NAME]
	ctx := r.Context()
	devices, err := getCommandsByDeviceName(ctx, dn, dbClient, deviceClient, configuration)
	if err != nil {
		httpErrorHandler.HandleManyVariants(
			w,
			err,
			[]errorconcept.ErrorConceptType{
				errorconcept.NewServiceClientHttpError(err),
				errorconcept.Device.NotFoundInDB,
			},
			errorconcept.Default.InternalServerError)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(&devices)
}

func restGetAllCommands(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	configuration *config.ConfigurationStruct,
	httpErrorHandler errorconcept.ErrorHandler) {

	ctx := r.Context()
	devices, err := getAllCommands(ctx, dbClient, deviceClient, configuration)
	if err != nil {
		httpErrorHandler.HandleManyVariants(
			w,
			err,
			[]errorconcept.ErrorConceptType{
				errorconcept.NewServiceClientHttpError(err),
				errorconcept.Database.NotFound,
			},
			errorconcept.Default.InternalServerError)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(devices)
}
