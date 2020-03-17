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
	"reflect"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/config"
	pwErrors "github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/gorilla/mux"
)

// creation endpoints
func restAddProvisionWatcher(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	defer r.Body.Close()
	var pw models.ProvisionWatcher
	var err error

	if err = json.NewDecoder(r.Body).Decode(&pw); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.JsonDecoding)
		return
	}

	// Check if the name exists
	if pw.Name == "" {
		errorHandler.Handle(
			w,
			errors.New("no name provided for new provision watcher"),
			errorconcept.Common.InvalidRequest_StatusBadRequest,
		)
		return
	}

	// Check if the device profile exists
	var profile models.DeviceProfile
	if pw.Profile.Id != "" {
		profile, err = dbClient.GetDeviceProfileById(pw.Profile.Id)
		// we don't want to fail if we haven't found anything at this point because we've yet to do name lookup
		if err != nil && err != db.ErrNotFound {
			errorHandler.Handle(w, err, errorconcept.Default.InternalServerError)
			return
		}
	}
	// we only want to do a lookup by name if ID does not exist or if ID does not exist in the DB
	if pw.Profile.Id == "" || err == db.ErrNotFound {
		if profile, err = dbClient.GetDeviceProfileByName(pw.Profile.Name); err != nil {
			errorHandler.HandleOneVariant(
				w,
				err,
				errorconcept.ProvisionWatcher.DeviceProfileNotFound_StatusConflict,
				errorconcept.Default.InternalServerError,
			)
			return
		}
	}
	pw.Profile = profile

	// Check if the device service exists
	var service models.DeviceService
	if pw.Service.Id != "" {
		service, err = dbClient.GetDeviceServiceById(pw.Service.Id)
		// we don't want to fail if we haven't found anything at this point because we've yet to do name lookup
		if err != nil && err != db.ErrNotFound {
			errorHandler.Handle(w, err, errorconcept.Default.ServiceUnavailable)
			return
		}
	}

	// we only want to do a lookup by name if ID does not exist or if ID does not exist in the DB
	if pw.Service.Id == "" || err == db.ErrNotFound {
		if service, err = dbClient.GetDeviceServiceByName(pw.Service.Name); err != nil {
			errorHandler.HandleOneVariant(
				w,
				err,
				errorconcept.ProvisionWatcher.DeviceServiceNotFound_StatusConflict,
				errorconcept.Default.ServiceUnavailable,
			)
			return
		}
	}
	pw.Service = service

	id, err := dbClient.AddProvisionWatcher(pw)
	if err != nil {
		errorHandler.HandleOneVariant(
			w,
			err,
			errorconcept.ProvisionWatcher.NotUnique,
			errorconcept.Default.ServiceUnavailable,
		)
		return
	}
	pw.Id = id

	// Notify Associates
	if err = notifyProvisionWatcherAssociates(pw, http.MethodPost, lc, dbClient); err != nil {
		errorHandler.Handle(w, err, errorconcept.Default.ServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(id))
}

// read endpoints
func restGetProvisionWatchers(
	w http.ResponseWriter,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	configuration *config.ConfigurationStruct) {

	res, err := dbClient.GetAllProvisionWatchers()
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusInternalServer)
		return
	}

	// Check the length
	if err = checkMaxLimit(len(res), configuration); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.LimitExceeded)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(&res)
}

func restGetProvisionWatcherById(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	var id = vars[ID]

	res, err := dbClient.GetProvisionWatcherById(id)
	if err != nil {
		errorHandler.HandleOneVariant(
			w,
			err,
			errorconcept.ProvisionWatcher.NotFoundById,
			errorconcept.Default.InternalServerError)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(res)
}

func restGetProvisionWatcherByName(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Default.BadRequest)
		return
	}

	res, err := dbClient.GetProvisionWatcherByName(n)
	if err != nil {
		errorHandler.HandleOneVariant(
			w,
			err,
			errorconcept.ProvisionWatcher.NotFoundByName,
			errorconcept.Default.InternalServerError)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(res)
}

func restGetProvisionWatchersByProfileId(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	var pid = vars[ID]

	// Check if the device profile exists
	if _, err := dbClient.GetDeviceProfileById(pid); err != nil {
		errorHandler.Handle(w, err, errorconcept.ProvisionWatcher.DeviceProfileNotFound_StatusNotFound)
		return
	}

	res, err := dbClient.GetProvisionWatchersByProfileId(pid)
	if err != nil {
		errorHandler.HandleOneVariant(
			w,
			err,
			errorconcept.ProvisionWatcher.NotFoundById,
			errorconcept.Common.RetrieveError_StatusInternalServer,
		)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(res)
}

func restGetProvisionWatchersByProfileName(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	pn, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the device profile exists
	dp, err := dbClient.GetDeviceProfileByName(pn)
	if err != nil {
		errorHandler.HandleOneVariant(
			w,
			err,
			errorconcept.ProvisionWatcher.DeviceProfileNotFound_StatusNotFound,
			errorconcept.Default.InternalServerError,
		)
		return
	}

	res, err := dbClient.GetProvisionWatchersByProfileId(dp.Id)
	if err != nil {
		errorHandler.HandleOneVariant(
			w,
			err,
			errorconcept.ProvisionWatcher.NotFoundById,
			errorconcept.Common.RetrieveError_StatusInternalServer,
		)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(res)
}

func restGetProvisionWatchersByServiceId(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	var sid = vars[ID]

	// Check if the device service exists
	if _, err := dbClient.GetDeviceServiceById(sid); err != nil {
		errorHandler.Handle(w, err, errorconcept.ProvisionWatcher.DeviceServiceNotFound_StatusNotFound)
		return
	}

	res, err := dbClient.GetProvisionWatchersByServiceId(sid)
	if err != nil {
		errorHandler.HandleOneVariant(
			w,
			err,
			errorconcept.ProvisionWatcher.NotFoundById,
			errorconcept.Common.RetrieveError_StatusInternalServer,
		)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(res)
}

func restGetProvisionWatchersByServiceName(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	sn, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the device service exists
	ds, err := dbClient.GetDeviceServiceByName(sn)
	if err != nil {
		errorHandler.HandleOneVariant(
			w,
			err,
			errorconcept.ProvisionWatcher.DeviceServiceNotFound_StatusNotFound,
			errorconcept.Default.InternalServerError,
		)
		return
	}

	// Get the provision watchers
	res, err := dbClient.GetProvisionWatchersByServiceId(ds.Id)
	if err != nil {
		errorHandler.HandleOneVariant(
			w,
			err,
			errorconcept.ProvisionWatcher.NotFoundById,
			errorconcept.Common.RetrieveError_StatusInternalServer,
		)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(res)
}

func restGetProvisionWatchersByIdentifier(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	k, err := url.QueryUnescape(vars[KEY])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}
	v, err := url.QueryUnescape(vars[VALUE])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	res, err := dbClient.GetProvisionWatchersByIdentifier(k, v)
	if err != nil {
		errorHandler.HandleOneVariant(
			w,
			err,
			errorconcept.ProvisionWatcher.RetrieveError_StatusNotFound,
			errorconcept.Common.RetrieveError_StatusInternalServer,
		)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(res)
}

// update endpoints
func restUpdateProvisionWatcher(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	defer r.Body.Close()
	var to models.ProvisionWatcher
	if err := json.NewDecoder(r.Body).Decode(&to); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusServiceUnavailable)
		return
	}

	from, err := getProvisionWatcher(to, dbClient)
	if err != nil {
		errorHandler.HandleManyVariants(
			w,
			err,
			[]errorconcept.ErrorConceptType{
				errorconcept.ProvisionWatcher.NotFoundByName,
				errorconcept.ProvisionWatcher.NameCollision,
			},
			errorconcept.Common.RetrieveError_StatusInternalServer,
		)
		return
	}

	// from is the object currently in the database; "the provision watcher we are updating *from*"
	// to is the target object state; "the provision watcher we are updating *to*"
	if to.Origin != 0 {
		from.Origin = to.Origin
	}

	if to.Name != "" {
		from.Name = to.Name
	}

	if to.Identifiers != nil {
		from.Identifiers = to.Identifiers
	}

	if to.BlockingIdentifiers != nil {
		from.BlockingIdentifiers = to.BlockingIdentifiers
	}

	if !reflect.DeepEqual(to.Profile, models.DeviceProfile{}) {
		from.Profile = to.Profile
	}

	if !reflect.DeepEqual(to.Service, models.DeviceService{}) {
		from.Service = to.Service
	}

	// always update admin state
	from.AdminState = to.AdminState

	if err := dbClient.UpdateProvisionWatcher(from); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.UpdateError_StatusInternalServer)
		return
	}

	if err := notifyProvisionWatcherAssociates(from, http.MethodPut, lc, dbClient); err != nil {
		errorHandler.Handle(w, err, errorconcept.Default.InternalServerError)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// getProvisionWatcher is a helper function that first attempts to lookup by ID, and, failing that,
// will attempt a lookup by name before finally bubbling the error up
func getProvisionWatcher(lookup models.ProvisionWatcher, dbClient interfaces.DBClient) (models.ProvisionWatcher, error) {
	byID, err := dbClient.GetProvisionWatcherById(lookup.Id)
	if err != nil && err != db.ErrNotFound { // ignore ErrNotFound, we can still do a name lookup
		return models.ProvisionWatcher{}, err
	}

	byName, err := dbClient.GetProvisionWatcherByName(lookup.Name)
	if err != nil {
		return models.ProvisionWatcher{}, err
	}

	// ensure that neither the lookup by name nor the lookup by ID have unwanted name collisions
	if namesCollide(byName, lookup) {
		return models.ProvisionWatcher{}, pwErrors.NewErrNameCollision(byName.Name, byName.Id, lookup.Id)
	} else if namesCollide(byName, byID) {
		return models.ProvisionWatcher{}, pwErrors.NewErrNameCollision(byName.Name, byName.Id, byID.Id)
	}

	// only use the result of the name lookup if we didn't get anything from the ID lookup
	if reflect.DeepEqual(byID, models.ProvisionWatcher{}) {
		return byName, nil
	}

	return byID, nil
}

// namesCollide consolidates the logic of determining whether provision watcher data that we are attempting to update
// will have either an ID or a name collision
func namesCollide(base models.ProvisionWatcher, update models.ProvisionWatcher) bool {
	// breaking this down piece by piece:
	// there is a collision if base and update have the same name ...
	//
	// ... only if the IDs don't match. If the IDs match, then we are correctly performing an update on an existing PW
	//
	// if the update PW has no ID, then we never say there is a collision and always assume that this is a valid update.
	// or, thinking about it differently, this clause is "update by name"
	return base.Name == update.Name && base.Id != update.Id && update.Id != ""
}

func restDeleteProvisionWatcherById(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	var id = vars[ID]

	// Check if the provision watcher exists
	pw, err := dbClient.GetProvisionWatcherById(id)
	if err != nil {
		errorHandler.HandleOneVariant(
			w,
			err,
			errorconcept.ProvisionWatcher.NotFoundById,
			errorconcept.Default.InternalServerError)
		return
	}

	if err = deleteProvisionWatcher(pw, w, lc, dbClient, errorHandler); err != nil {
		errorHandler.Handle(w, err, errorconcept.ProvisionWatcher.DeleteError_StatusInternalServer)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restDeleteProvisionWatcherByName(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the provision watcher exists
	pw, err := dbClient.GetProvisionWatcherByName(n)
	if err != nil {
		errorHandler.HandleOneVariant(
			w,
			err,
			errorconcept.ProvisionWatcher.NotFoundByName,
			errorconcept.Default.InternalServerError)
		return
	}

	if err = deleteProvisionWatcher(pw, w, lc, dbClient, errorHandler); err != nil {
		errorHandler.Handle(w, err, errorconcept.ProvisionWatcher.DeleteError_StatusInternalServer)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// delete endpoints
func deleteProvisionWatcher(
	pw models.ProvisionWatcher,
	w http.ResponseWriter,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) error {

	if err := dbClient.DeleteProvisionWatcherById(pw.Id); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.DeleteError)
		return err
	}

	err := notifyProvisionWatcherAssociates(pw, http.MethodDelete, lc, dbClient)
	if err != nil {
		lc.Error("Problem notifying associated device services to provision watcher: " + err.Error())
		return err
	}

	return nil
}

// notifyProvisionWatcherAssociates triggers the callbacks in the device service attached to this provision watcher.
func notifyProvisionWatcherAssociates(
	pw models.ProvisionWatcher,
	action string,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) error {

	// Get the device service for the provision watcher
	ds, err := dbClient.GetDeviceServiceById(pw.Service.Id)
	if err != nil {
		// try by name
		ds, err = dbClient.GetDeviceServiceByName(pw.Service.Name)
		if err != nil {
			return err
		}
	}

	// notify the device service
	err = notifyAssociates([]models.DeviceService{ds}, pw.Id, action, models.PROVISIONWATCHER, lc)
	if err != nil {
		return err
	}

	return nil
}
