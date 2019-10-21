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
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"
)

// Returns 200, 400, 404, 423, 500 according to the RAML
func restGetDeviceCommandByCommandID(w http.ResponseWriter, r *http.Request) {
	issueDeviceCommand(w, r, false)
}

// Returns 200, 400, 404, 423, 500 according to the RAML
func restPutDeviceCommandByCommandID(w http.ResponseWriter, r *http.Request) {
	issueDeviceCommand(w, r, true)
}

// Returns 200, 400, 404, 423, 500 according to the RAML
func issueDeviceCommand(w http.ResponseWriter, r *http.Request, isPutCommand bool) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	did := vars[ID]
	cid := vars[COMMANDID]
	b, err := ioutil.ReadAll(r.Body)
	if b == nil && err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	ctx := r.Context()
	body, err := commandByDeviceID(did, cid, string(b), r.URL.RawQuery, isPutCommand, ctx)
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

	w.WriteHeader(http.StatusOK)
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.Write([]byte(body))
}

// Returns 200, 400, 404, 423, 500 according to the RAML
func restGetDeviceCommandByNames(w http.ResponseWriter, r *http.Request) {
	issueDeviceCommandByNames(w, r, false)
}

// Returns 200, 400, 404, 423, 500 according to the RAML
func restPutDeviceCommandByNames(w http.ResponseWriter, r *http.Request) {
	issueDeviceCommandByNames(w, r, true)
}

// Returns 200, 400, 404, 423, 500 according to the RAML
func issueDeviceCommandByNames(w http.ResponseWriter, r *http.Request, isPutCommand bool) {
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
	body, err := commandByNames(dn, cn, string(b), r.URL.RawQuery, isPutCommand, ctx)

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

	w.WriteHeader(http.StatusOK)
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.Write([]byte(body))
}

// Returns 200, 404, 500 according to the RAML
func restGetCommandsByDeviceID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	did := vars[ID]
	ctx := r.Context()
	device, err := getCommandsByDeviceID(did, ctx)
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
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(&device)
}

// Returns 200, 404, 500 according to the RAML
func restGetCommandsByDeviceName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dn := vars[NAME]
	ctx := r.Context()
	devices, err := getCommandsByDeviceName(dn, ctx)
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
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(&devices)
}

// Returns 200 and 500 according to the RAML
func restGetAllCommands(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	devices, err := getCommands(ctx)
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

	w.WriteHeader(http.StatusOK)
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(devices)
}
