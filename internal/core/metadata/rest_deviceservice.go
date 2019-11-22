/*******************************************************************************
 * Copyright 2019 Dell Inc.
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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device_service"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/gorilla/mux"
)

func restGetAllDeviceServices(
	w http.ResponseWriter,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	op := device_service.NewDeviceServiceLoadAll(Configuration.Service, dbClient, loggingClient)
	services, err := op.Execute()
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Common.LimitExceeded, errorconcept.Default.InternalServerError)
		return
	}
	pkg.Encode(services, w, loggingClient)
}

func restAddDeviceService(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	defer r.Body.Close()
	var ds models.DeviceService
	err := json.NewDecoder(r.Body).Decode(&ds)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Addressable Check
	// No ID or Name given for addressable
	if ds.Addressable.Id == "" && ds.Addressable.Name == "" {
		errorHandler.Handle(w, err, errorconcept.DeviceService.EmptyAddressable)
		return
	}

	// First try by name
	addressable, err := dbClient.GetAddressableByName(ds.Addressable.Name)
	if err != nil && err == db.ErrNotFound && ds.Addressable.Id != "" {
		addressable, err = dbClient.GetAddressableById(ds.Addressable.Id)
	}
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.DeviceService.AddressableNotFound, errorconcept.Default.InternalServerError)
		return
	}
	ds.Addressable = addressable

	// Add the device service
	if ds.Id, err = dbClient.AddDeviceService(ds); err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.DeviceService.NotUnique, errorconcept.Default.InternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(ds.Id))
}

func restUpdateDeviceService(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	defer r.Body.Close()
	var from models.DeviceService
	err := json.NewDecoder(r.Body).Decode(&from)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the device service exists and get it
	var to models.DeviceService
	// Try by ID
	if from.Id != "" {
		to, err = dbClient.GetDeviceServiceById(from.Id)
	}
	if from.Id == "" || err != nil {
		// Try by Name
		if to, err = dbClient.GetDeviceServiceByName(from.Name); err != nil {
			errorHandler.Handle(w, err, errorconcept.DeviceService.NotFound)
			return
		}
	}

	if err = updateDeviceServiceFields(from, &to, w, dbClient, errorHandler); err != nil {
		loggingClient.Error(err.Error())
		return
	}

	if err := dbClient.UpdateDeviceService(to); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.UpdateError_StatusInternalServer)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the relevant device service fields
func updateDeviceServiceFields(
	from models.DeviceService,
	to *models.DeviceService,
	w http.ResponseWriter,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) error {

	// Use .String() to compare empty structs (not ideal, but there is no .equals method)
	if (from.Addressable.String() != models.Addressable{}.String()) {
		var addr models.Addressable
		var err error
		if from.Addressable.Id != "" {
			addr, err = dbClient.GetAddressableById(from.Addressable.Id)
		}
		if from.Addressable.Id == "" || err != nil {
			addr, err = dbClient.GetAddressableByName(from.Addressable.Name)
			if err != nil {
				return err
			}
		}
		to.Addressable = addr
	}

	to.AdminState = from.AdminState
	if from.Description != "" {
		to.Description = from.Description
	}
	if from.Labels != nil {
		to.Labels = from.Labels
	}
	if from.LastConnected != 0 {
		to.LastConnected = from.LastConnected
	}
	if from.LastReported != 0 {
		to.LastReported = from.LastReported
	}
	if from.Name != "" {
		to.Name = from.Name

		// Check if the new name is unique
		checkDS, err := dbClient.GetDeviceServiceByName(from.Name)
		if err != nil {
			errorHandler.HandleOneVariant(w, err, errorconcept.NewDeviceServiceDuplicate(checkDS.Id, to.Id), errorconcept.Default.ServiceUnavailable)
		}
	}

	to.OperatingState = from.OperatingState
	if from.Origin != 0 {
		to.Origin = from.Origin
	}

	return nil
}

func restGetServiceByAddressableName(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	an, err := url.QueryUnescape(vars[ADDRESSABLENAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	op := device_service.NewDeviceServiceLoadByAddressableName(an, dbClient)
	res, err := op.Execute()
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Common.ItemNotFound, errorconcept.Default.InternalServerError)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(res)
}

func restGetServiceByAddressableId(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	var sid = vars[ADDRESSABLEID]

	op := device_service.NewDeviceServiceLoadByAddressableID(sid, dbClient)
	res, err := op.Execute()
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Common.ItemNotFound, errorconcept.Default.InternalServerError)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(res)
}

func restGetServiceWithLabel(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	l, err := url.QueryUnescape(vars[LABEL])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	res := make([]models.DeviceService, 0)
	if res, err = dbClient.GetDeviceServicesWithLabel(l); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusInternalServer)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(res)
}

func restGetServiceByName(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	dn, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	op := device_service.NewDeviceServiceLoadByName(dn, dbClient)
	res, err := op.Execute()
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Common.ItemNotFound, errorconcept.Default.InternalServerError)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(res)
}

func restDeleteServiceById(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	var id string = vars[ID]

	// Check if the device service exists and get it
	ds, err := dbClient.GetDeviceServiceById(id)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.DeviceService.NotFound, errorconcept.Default.InternalServerError)
		return
	}

	ctx := r.Context()
	if err = deleteDeviceService(ds, w, ctx, loggingClient, dbClient, errorHandler); err != nil {
		loggingClient.Error(err.Error())
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.Write([]byte("true"))
}

func restDeleteServiceByName(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the device service exists
	ds, err := dbClient.GetDeviceServiceByName(n)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.DeviceService.NotFound, errorconcept.Default.InternalServerError)
		return
	}

	ctx := r.Context()
	// Delete the device service
	if err = deleteDeviceService(ds, w, ctx, loggingClient, dbClient, errorHandler); err != nil {
		loggingClient.Error(err.Error())
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Delete the device service
// Delete the associated devices
// Delete the associated provision watchers
func deleteDeviceService(
	ds models.DeviceService,
	w http.ResponseWriter,
	ctx context.Context,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) error {

	// Delete the associated devices
	devices, err := dbClient.GetDevicesByServiceId(ds.Id)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusServiceUnavailable)
		return err
	}
	for _, device := range devices {
		if err = deleteDevice(device, w, ctx, loggingClient, dbClient, errorHandler); err != nil {
			return err
		}
	}

	// Delete the associated provision watchers
	watchers, err := dbClient.GetProvisionWatchersByServiceId(ds.Id)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusServiceUnavailable)
		return err
	}
	for _, watcher := range watchers {
		if err = deleteProvisionWatcher(watcher, w, loggingClient, dbClient, errorHandler); err != nil {
			return err
		}
	}

	// Delete the device service
	if err = dbClient.DeleteDeviceServiceById(ds.Id); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.DeleteError)
		return err
	}

	return nil
}

func restUpdateServiceLastConnectedById(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	var id string = vars[ID]
	var vlc string = vars[LASTCONNECTED]
	lc, err := strconv.ParseInt(vlc, 10, 64)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the device service exists
	ds, err := dbClient.GetDeviceServiceById(id)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.DeviceService.NotFound, errorconcept.Default.InternalServerError)
		return
	}

	if err = updateServiceLastConnected(ds, lc, w, dbClient, errorHandler); err != nil {
		loggingClient.Error(err.Error())
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restUpdateServiceLastConnectedByName(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}
	var vlc string = vars[LASTCONNECTED]
	lc, err := strconv.ParseInt(vlc, 10, 64)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.DeviceService.InvalidRequest_StatusInternalServer)
		return
	}

	// Check if the device service exists
	ds, err := dbClient.GetDeviceServiceByName(n)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.DeviceService.NotFound, errorconcept.Default.InternalServerError)
		return
	}

	// Update last connected
	if err = updateServiceLastConnected(ds, lc, w, dbClient, errorHandler); err != nil {
		loggingClient.Error(err.Error())
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the last connected value of the device service
func updateServiceLastConnected(
	ds models.DeviceService,
	lc int64,
	w http.ResponseWriter,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) error {

	ds.LastConnected = lc

	if err := dbClient.UpdateDeviceService(ds); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.UpdateError_StatusServiceUnavailable)
		return err
	}

	return nil
}

func restGetServiceById(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	var did = vars[ID]

	op := device_service.NewDeviceServiceLoadById(did, dbClient)
	res, err := op.Execute()
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Common.ItemNotFound, errorconcept.Default.InternalServerError)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(res)
}

func restUpdateServiceOpStateById(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	var id = vars[ID]
	var os = vars[OPSTATE]

	// Check the OpState
	newOs, f := models.GetOperatingState(os)
	if !f {
		err := errors.New("Invalid State: " + os + " Must be 'ENABLED' or 'DISABLED'")
		errorHandler.Handle(w, err, errorconcept.DeviceService.InvalidState)
		return
	}

	op := device_service.NewUpdateOpStateByIdExecutor(id, newOs, dbClient)
	if err := op.Execute(); err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Common.ItemNotFound, errorconcept.Default.InternalServerError)
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restUpdateServiceOpStateByName(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}
	var os = vars[OPSTATE]

	// Check the OpState
	newOs, f := models.GetOperatingState(os)
	if !f {
		err = errors.New("Invalid State: " + os + " Must be 'ENABLED' or 'DISABLED'")
		errorHandler.Handle(w, err, errorconcept.DeviceService.InvalidState)
		return
	}

	op := device_service.NewUpdateOpStateByNameExecutor(n, newOs, dbClient)
	if err := op.Execute(); err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Common.ItemNotFound, errorconcept.Default.InternalServerError)
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restUpdateServiceAdminStateById(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	var id = vars[ID]
	var as = vars[ADMINSTATE]

	// Check the admin state
	newAs, f := models.GetAdminState(as)
	if !f {
		err := errors.New("Invalid state: " + as + " Must be 'LOCKED' or 'UNLOCKED'")
		errorHandler.Handle(w, err, errorconcept.DeviceService.InvalidState)
		return
	}

	op := device_service.NewUpdateAdminStateByIdExecutor(id, newAs, dbClient)
	if err := op.Execute(); err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Common.ItemNotFound, errorconcept.Default.InternalServerError)
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restUpdateServiceAdminStateByName(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}
	var as = vars[ADMINSTATE]

	// Check the admin state
	newAs, f := models.GetAdminState(as)
	if !f {
		err := errors.New("Invalid state: " + as + " Must be 'LOCKED' or 'UNLOCKED'")
		errorHandler.Handle(w, err, errorconcept.DeviceService.InvalidState)
		return
	}

	op := device_service.NewUpdateAdminStateByNameExecutor(n, newAs, dbClient)
	if err := op.Execute(); err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Common.ItemNotFound, errorconcept.Default.InternalServerError)
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restUpdateServiceLastReportedById(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	var id string = vars[ID]
	var vlr string = vars[LASTREPORTED]
	lr, err := strconv.ParseInt(vlr, 10, 64)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the device service exists
	ds, err := dbClient.GetDeviceServiceById(id)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.DeviceService.NotFound, errorconcept.Default.InternalServerError)
		return
	}

	if err = updateServiceLastReported(ds, lr, w, dbClient, errorHandler); err != nil {
		loggingClient.Error(err.Error())
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restUpdateServiceLastReportedByName(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}
	var vlr string = vars[LASTREPORTED]
	lr, err := strconv.ParseInt(vlr, 10, 64)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the device service exists
	ds, err := dbClient.GetDeviceServiceByName(n)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.DeviceService.NotFound, errorconcept.Default.InternalServerError)
		return
	}

	if err = updateServiceLastReported(ds, lr, w, dbClient, errorHandler); err != nil {
		loggingClient.Error(err.Error())
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the last reported value for the device service
func updateServiceLastReported(
	ds models.DeviceService,
	lr int64,
	w http.ResponseWriter,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) error {

	ds.LastReported = lr
	if err := dbClient.UpdateDeviceService(ds); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.UpdateError_StatusServiceUnavailable)
		return err
	}

	return nil
}

// Notify associates (associated device services)
// This function is called when an object changes in metadata
func notifyAssociates(
	deviceServices []models.DeviceService,
	id string,
	action string,
	actionType models.ActionType,
	loggingClient logger.LoggingClient) error {

	for _, ds := range deviceServices {
		if err := callback(ds, id, action, actionType, loggingClient); err != nil {
			return err
		}
	}

	return nil
}

// Make the callback for the device service
func callback(
	service models.DeviceService,
	id string,
	action string,
	actionType models.ActionType,
	loggingClient logger.LoggingClient) error {

	client := &http.Client{}
	url := service.Addressable.GetCallbackURL()
	if len(url) > 0 {
		body, err := getBody(id, actionType)
		if err != nil {
			return err
		}
		req, err := http.NewRequest(string(action), url, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Add(clients.ContentType, clients.ContentTypeJSON)

		go makeRequest(client, req, loggingClient)
	} else {
		loggingClient.Info("callback::no addressable for " + service.Name)
	}
	return nil
}

// Asynchronous call
func makeRequest(client *http.Client, req *http.Request, loggingClient logger.LoggingClient) {
	// Make the request
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		resp.Close = true
	} else {
		loggingClient.Error(err.Error())
	}
}

// Turn the ID and ActionType into the JSON body that will be passed
func getBody(id string, actionType models.ActionType) ([]byte, error) {
	return json.Marshal(models.CallbackAlert{ActionType: actionType, Id: id})
}
