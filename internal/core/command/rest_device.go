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

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/core/command/config"
	"github.com/edgexfoundry/edgex-go/internal/core/command/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"

	"github.com/gorilla/mux"
)

func restGetDeviceCommandByCommandID(
	w http.ResponseWriter,
	originalRequest *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	httpErrorHandler errorconcept.ErrorHandler,
	httpCaller internal.HttpCaller) {

	issueDeviceCommand(w, originalRequest, lc, dbClient, deviceClient, httpErrorHandler, httpCaller)
}

func restPutDeviceCommandByCommandID(
	w http.ResponseWriter,
	originalRequest *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	httpErrorHandler errorconcept.ErrorHandler,
	httpCaller internal.HttpCaller) {

	issueDeviceCommand(w, originalRequest, lc, dbClient, deviceClient, httpErrorHandler, httpCaller)
}

func issueDeviceCommand(
	w http.ResponseWriter,
	originalRequest *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	httpErrorHandler errorconcept.ErrorHandler,
	httpCaller internal.HttpCaller) {

	defer originalRequest.Body.Close()

	b, err := ioutil.ReadAll(originalRequest.Body)
	if b == nil && err != nil {
		lc.Error(err.Error())
		return
	}

	deviceServiceResponse, deviceServiceResponseBody, err := executeCommandByDeviceID(
		originalRequest,
		string(b),
		lc,
		dbClient,
		deviceClient,
		httpCaller)

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

	// Extract the headers from the device service's response, so
	// we can propagate it to functions upward in the call chain.
	headers := map[string]string{"Content-Type": ""}
	m := make(map[string][]string)
	m = deviceServiceResponse.Header
	for name, values := range m {
		for _, value := range values {
			headers[name] = value
		}
	}

	// Set the returned header Content-type based on header Content-type received in
	// the Device Service request (No need to inspect it).
	w.Header().Set(clients.ContentType, headers[clients.ContentType])
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(deviceServiceResponseBody))
}

func restGetDeviceCommandByNames(
	w http.ResponseWriter,
	originalRequest *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	httpErrorHandler errorconcept.ErrorHandler,
	httpCaller internal.HttpCaller) {

	issueDeviceCommandByNames(w, originalRequest, lc, dbClient, deviceClient, httpErrorHandler, httpCaller)
}

func restPutDeviceCommandByNames(
	w http.ResponseWriter,
	originalRequest *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	httpErrorHandler errorconcept.ErrorHandler,
	httpCaller internal.HttpCaller) {

	issueDeviceCommandByNames(w, originalRequest, lc, dbClient, deviceClient, httpErrorHandler, httpCaller)
}

func issueDeviceCommandByNames(
	w http.ResponseWriter,
	originalRequest *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	httpErrorHandler errorconcept.ErrorHandler,
	httpCaller internal.HttpCaller) {

	defer originalRequest.Body.Close()

	vars := mux.Vars(originalRequest)
	dn := vars[NAME]
	cn := vars[COMMANDNAME]

	ctx := originalRequest.Context()

	b, err := ioutil.ReadAll(originalRequest.Body)
	if b == nil && err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	deviceServiceResponse, deviceServiceResponseBody, err := executeCommandByName(
		originalRequest,
		ctx,
		dn,
		cn,
		string(b),
		lc,
		dbClient,
		deviceClient,
		httpCaller)

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

	// Extract the headers from the device service's response, so
	// we can propagate it to functions upward in the call chain.
	headers := map[string]string{"Content-Type": ""}
	m := make(map[string][]string)
	m = deviceServiceResponse.Header
	for name, values := range m {
		for _, value := range values {
			headers[name] = value
		}
	}

	// Set the returned header Content-type based on header Content-type received in
	// the Device Service request (No need to inspect it).
	w.Header().Set(clients.ContentType, headers[clients.ContentType])
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(deviceServiceResponseBody))
}

func restGetCommandsByDeviceID(
	w http.ResponseWriter,
	originalRequest *http.Request,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	configuration *config.ConfigurationStruct,
	httpErrorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(originalRequest)
	did := vars[ID]
	ctx := originalRequest.Context()
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
	originalRequest *http.Request,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	configuration *config.ConfigurationStruct,
	httpErrorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(originalRequest)
	dn := vars[NAME]
	ctx := originalRequest.Context()
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

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(&devices)
}

func restGetAllCommands(
	w http.ResponseWriter,
	originalRequest *http.Request,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	configuration *config.ConfigurationStruct,
	httpErrorHandler errorconcept.ErrorHandler) {

	ctx := originalRequest.Context()
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
