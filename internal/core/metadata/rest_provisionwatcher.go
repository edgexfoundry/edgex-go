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
	"errors"
	"net/http"
	"net/url"

	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/gorilla/mux"
)

func restGetProvisionWatchers(
	w http.ResponseWriter) {

	res, err := dbClient.GetAllProvisionWatchers()
	if err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusServiceUnavailable)
		return
	}

	// Check the length
	if err = checkMaxLimit(len(res)); err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Common.LimitExceeded)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(&res)
}

func restDeleteProvisionWatcherById(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient) {

	vars := mux.Vars(r)
	var id string = vars[ID]

	// Check if the provision watcher exists
	pw, err := dbClient.GetProvisionWatcherById(id)
	if err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.ProvisionWatcher.NotFoundById)
		return
	}

	err = deleteProvisionWatcher(pw, w, loggingClient)
	if err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.ProvisionWatcher.DeleteError_StatusInternalServer)
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.Write([]byte("true"))
}

func restDeleteProvisionWatcherByName(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient) {

	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the provision watcher exists
	pw, err := dbClient.GetProvisionWatcherByName(n)
	if err != nil {
		httpErrorHandler.HandleOneVariant(w, err, errorconcept.ProvisionWatcher.NotFoundByName, errorconcept.Default.InternalServerError)
		return
	}

	if err = deleteProvisionWatcher(pw, w, loggingClient); err != nil {
		loggingClient.Error("Problem deleting provision watcher: " + err.Error())
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Delete the provision watcher
func deleteProvisionWatcher(
	pw models.ProvisionWatcher,
	w http.ResponseWriter,
	loggingClient logger.LoggingClient) error {

	if err := dbClient.DeleteProvisionWatcherById(pw.Id); err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Common.DeleteError)
		return err
	}

	if err := notifyProvisionWatcherAssociates(pw, http.MethodDelete, loggingClient); err != nil {
		loggingClient.Error("Problem notifying associated device services to provision watcher: " + err.Error())
	}

	return nil
}

func restGetProvisionWatcherById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var id string = vars[ID]

	res, err := dbClient.GetProvisionWatcherById(id)
	if err != nil {
		httpErrorHandler.HandleOneVariant(w, err, errorconcept.ProvisionWatcher.NotFoundById, errorconcept.Default.InternalServerError)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(res)
}

func restGetProvisionWatcherByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Default.BadRequest)
		return
	}

	res, err := dbClient.GetProvisionWatcherByName(n)
	if err != nil {
		httpErrorHandler.HandleOneVariant(w, err, errorconcept.ProvisionWatcher.NotFoundByName, errorconcept.Default.InternalServerError)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(res)
}

func restGetProvisionWatchersByProfileId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var pid string = vars[ID]

	// Check if the device profile exists
	if _, err := dbClient.GetDeviceProfileById(pid); err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.ProvisionWatcher.DeviceProfileNotFound_StatusNotFound)
		return
	}

	res, err := dbClient.GetProvisionWatchersByProfileId(pid)
	if err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusInternalServer)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(res)
}

func restGetProvisionWatchersByProfileName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pn, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the device profile exists
	dp, err := dbClient.GetDeviceProfileByName(pn)
	if err != nil {
		httpErrorHandler.HandleOneVariant(w, err, errorconcept.ProvisionWatcher.DeviceProfileNotFound_StatusNotFound, errorconcept.Default.InternalServerError)
		return
	}

	res, err := dbClient.GetProvisionWatchersByProfileId(dp.Id)
	if err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusInternalServer)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(res)
}

func restGetProvisionWatchersByServiceId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var sid string = vars[ID]

	// Check if the device service exists
	if _, err := dbClient.GetDeviceServiceById(sid); err != nil {
		httpErrorHandler.Handle(w, errors.New("Device Service not found"), errorconcept.Default.NotFound)
		return
	}

	res, err := dbClient.GetProvisionWatchersByServiceId(sid)
	if err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusInternalServer)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(res)
}

func restGetProvisionWatchersByServiceName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sn, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the device service exists
	ds, err := dbClient.GetDeviceServiceByName(sn)
	if err != nil {
		httpErrorHandler.HandleOneVariant(w, err, errorconcept.ProvisionWatcher.DeviceServiceNotFound_StatusNotFound, errorconcept.Default.InternalServerError)
		return
	}

	// Get the provision watchers
	res, err := dbClient.GetProvisionWatchersByServiceId(ds.Id)
	if err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.ProvisionWatcher.RetrieveError_StatusNotFound)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(res)
}

func restGetProvisionWatchersByIdentifier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	k, err := url.QueryUnescape(vars[KEY])
	if err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}
	v, err := url.QueryUnescape(vars[VALUE])
	if err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	res, err := dbClient.GetProvisionWatchersByIdentifier(k, v)
	if err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusInternalServer)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(res)
}

func restAddProvisionWatcher(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient) {

	defer r.Body.Close()
	var pw models.ProvisionWatcher
	var err error

	if err = json.NewDecoder(r.Body).Decode(&pw); err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusServiceUnavailable)
		return
	}

	// Check if the name exists
	if pw.Name == "" {
		httpErrorHandler.HandleOneVariant(w, errors.New("No name provided for new provision watcher"), nil, errorconcept.Default.Conflict)
		return
	}

	// Check if the device profile exists
	// Try by ID
	var profile models.DeviceProfile
	if pw.Profile.Id != "" {
		profile, err = dbClient.GetDeviceProfileById(pw.Profile.Id)
	}
	if pw.Profile.Id == "" || err != nil {
		// Try by name
		if profile, err = dbClient.GetDeviceProfileByName(pw.Profile.Name); err != nil {
			httpErrorHandler.HandleOneVariant(w, err, errorconcept.ProvisionWatcher.DeviceProfileNotFound_StatusConflict, errorconcept.Default.ServiceUnavailable)
			return
		}
	}
	pw.Profile = profile

	// Check if the device service exists
	// Try by ID
	var service models.DeviceService
	if pw.Service.Id != "" {
		service, err = dbClient.GetDeviceServiceById(pw.Service.Id)
	}
	if pw.Service.Id == "" || err != nil {
		// Try by name
		if service, err = dbClient.GetDeviceServiceByName(pw.Service.Name); err != nil {
			httpErrorHandler.HandleOneVariant(w, err, errorconcept.ProvisionWatcher.DeviceServiceNotFound_StatusConflict, errorconcept.Default.ServiceUnavailable)
			return
		}
	}
	pw.Service = service

	id, err := dbClient.AddProvisionWatcher(pw)
	if err != nil {
		httpErrorHandler.HandleOneVariant(w, err, errorconcept.ProvisionWatcher.NotUnique, errorconcept.Default.ServiceUnavailable)
		return
	}

	// Notify Associates
	if err = notifyProvisionWatcherAssociates(pw, http.MethodPost, loggingClient); err != nil {
		loggingClient.Error("Problem with notifying associating device services for the provision watcher: " + err.Error())
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(id))
}

// Update the provision watcher object
// ID is used first for identification, then name
// The service and profile cannot be updated
func restUpdateProvisionWatcher(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient) {

	defer r.Body.Close()
	var from models.ProvisionWatcher
	if err := json.NewDecoder(r.Body).Decode(&from); err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusServiceUnavailable)
		return
	}

	// Check if the provision watcher exists
	// Try by ID
	to, err := dbClient.GetProvisionWatcherById(from.Id)
	if err != nil {
		// Try by name
		if to, err = dbClient.GetProvisionWatcherByName(from.Name); err != nil {
			httpErrorHandler.HandleOneVariant(w, err, errorconcept.ProvisionWatcher.NotFoundByName, errorconcept.Common.RetrieveError_StatusServiceUnavailable)
			return
		}
	}

	if err := updateProvisionWatcherFields(from, &to, w); err != nil {
		loggingClient.Error("Problem updating provision watcher: " + err.Error())
		return
	}

	if err := dbClient.UpdateProvisionWatcher(to); err != nil {
		httpErrorHandler.Handle(w, err, errorconcept.Common.UpdateError_StatusServiceUnavailable)
		return
	}

	// Notify Associates
	if err := notifyProvisionWatcherAssociates(to, http.MethodPut, loggingClient); err != nil {
		loggingClient.Error("Problem notifying associated device services for provision watcher: " + err.Error())
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the relevant fields of the provision watcher
func updateProvisionWatcherFields(from models.ProvisionWatcher, to *models.ProvisionWatcher, w http.ResponseWriter) error {
	if from.Identifiers != nil {
		to.Identifiers = from.Identifiers
	}
	if from.Origin != 0 {
		to.Origin = from.Origin
	}
	if from.Name != "" {
		// Check that the name is unique
		checkPW, err := dbClient.GetProvisionWatcherByName(from.Name)
		if err != nil {
			// DuplicateProvisionWatcherErrorConcept will evaluate to true if the ID is a duplicate
			httpErrorHandler.HandleOneVariant(w, err, errorconcept.NewProvisionWatcherDuplicateErrorConcept(checkPW.Id, to.Id), errorconcept.Default.ServiceUnavailable)
		}
		to.Name = from.Name
	}

	return nil
}

// Notify the associated device services for the provision watcher
func notifyProvisionWatcherAssociates(
	pw models.ProvisionWatcher,
	action string,
	loggingClient logger.LoggingClient) error {

	// Get the device service for the provision watcher
	ds, err := dbClient.GetDeviceServiceById(pw.Service.Id)
	if err != nil {
		return err
	}

	// Notify the service
	err = notifyAssociates([]models.DeviceService{ds}, pw.Id, action, models.PROVISIONWATCHER, loggingClient)
	if err != nil {
		return err
	}

	return nil
}
