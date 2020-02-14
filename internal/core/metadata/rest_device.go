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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/config"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/notifications"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-core-contracts/requests/states/admin"
	"github.com/edgexfoundry/go-mod-core-contracts/requests/states/operating"

	"github.com/gorilla/mux"
)

func restGetAllDevices(
	w http.ResponseWriter,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	configuration *config.ConfigurationStruct) {

	op := device.NewDeviceLoadAll(configuration.Service, dbClient, lc)
	devices, err := op.Execute()
	if err != nil {
		errorHandler.HandleOneVariant(
			w,
			err,
			errorconcept.Common.LimitExceeded,
			errorconcept.Default.InternalServerError)
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(&devices)
}

// Post a new device
// Attached objects (Addressable, Profile, Service) are referenced by ID or name
// 409 conflict if any of the attached items can't be found by ID or name
// Ignore everything else from the attached objects
func restAddNewDevice(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) {

	defer r.Body.Close()

	var d models.Device
	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		errorHandler.HandleOneVariant(
			w,
			err,
			errorconcept.Common.ContractInvalid_StatusBadRequest,
			errorconcept.Default.InternalServerError)
		return
	}

	ctx := r.Context()
	// The following requester instance is necessary because we will be making an HTTP call to the device service
	// associated with the new device in the Notifier below. There is no device service client. Additionally, the
	// requester interface should be mocked for unit testability and so is injected into the Notifier.
	requester, err := device.NewRequester(device.Http, lc, ctx)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Device.RequesterError)
		return
	}

	ch := make(chan device.DeviceEvent)
	defer close(ch)

	notifier := device.NewNotifier(ch, nc, configuration.Notifications, dbClient, requester, lc, ctx)
	go notifier.Execute()

	op := device.NewAddDevice(ch, dbClient, d)
	newId, err := op.Execute()
	if err != nil {
		errorHandler.HandleManyVariants(
			w,
			err,
			[]errorconcept.ErrorConceptType{
				errorconcept.Common.DuplicateName,
				errorconcept.Common.ItemNotFound,
			},
			errorconcept.Default.InternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(newId))
}

// Update the device
// Use ID to identify device first, then name
// Can't create new Device Services/Profiles with a PUT, but you can reference another one
func restUpdateDevice(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) {

	defer r.Body.Close()
	var rd models.Device
	err := json.NewDecoder(r.Body).Decode(&rd)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusServiceUnavailable)
		return
	}

	ch := make(chan device.DeviceEvent)
	defer close(ch)

	ctx := r.Context()

	requester, err := device.NewRequester(device.Http, lc, ctx)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Device.RequesterError)
		return
	}

	notifier := device.NewNotifier(ch, nc, configuration.Notifications, dbClient, requester, lc, ctx)
	go notifier.Execute()

	op := device.NewUpdateDevice(ch, dbClient, rd, lc)
	err = op.Execute()

	if err != nil {
		errorHandler.HandleManyVariants(
			w,
			err,
			[]errorconcept.ErrorConceptType{
				errorconcept.Common.DuplicateName,
				errorconcept.Common.ItemNotFound,
			},
			errorconcept.Default.InternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restGetDevicesWithLabel(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	label, err := url.QueryUnescape(vars[LABEL])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	res, err := dbClient.GetDevicesWithLabel(label)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusInternalServer)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(res)
}

func restGetDeviceByProfileId(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	var pid = vars[PROFILEID]

	// Check if the device profile exists
	_, err := dbClient.GetDeviceProfileById(pid)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.InternalServerError)
		return
	}

	res, err := dbClient.GetDevicesByProfileId(pid)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusInternalServer)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(res)
}

func restGetDeviceByServiceId(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	var sid = vars[SERVICEID]

	// Check if the device service exists
	_, err := dbClient.GetDeviceServiceById(sid)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.ServiceUnavailable)
		return
	}

	res, err := dbClient.GetDevicesByServiceId(sid)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusServiceUnavailable)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(res)
}

// If the result array is empty, don't return http.NotFound, just return empty array
func restGetDeviceByServiceName(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	sn, err := url.QueryUnescape(vars[SERVICENAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusServiceUnavailable)
		return
	}

	// Check if the device service exists
	ds, err := dbClient.GetDeviceServiceByName(sn)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.ServiceUnavailable)
		return
	}

	// Find devices by service ID now that you have the Service object (and therefor the ID)
	res, err := dbClient.GetDevicesByServiceId(ds.Id)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusServiceUnavailable)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(res)
}

func restGetDeviceByProfileName(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	pn, err := url.QueryUnescape(vars[PROFILENAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusServiceUnavailable)
		return
	}

	// Check if the device profile exists
	dp, err := dbClient.GetDeviceProfileByName(pn)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.ServiceUnavailable)
		return
	}

	// Use profile ID now that you have the profile object
	res, err := dbClient.GetDevicesByProfileId(dp.Id)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusServiceUnavailable)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(res)
}

func restGetDeviceById(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	var did = vars[ID]

	res, err := dbClient.GetDeviceById(did)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.BadRequest)
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(res)
}

// restCheckForDevice looks for a device using both its name and device, in that order. If found,
// it is returned as a JSON encoded string.
func restCheckForDevice(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	token := vars[ID] // referring to this as "token" for now since the source variable is double purposed

	// Check for name first since we're using that meaning by default.
	dev, err := dbClient.GetDeviceByName(token)
	if err != nil {
		if err != db.ErrNotFound {
			errorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusInternalServer)
			return
		} else {
			lc.Debug(fmt.Sprintf("device %s %v", token, err))
		}
	}

	// If lookup by name failed, see if we were passed the ID
	if len(dev.Name) == 0 {
		if dev, err = dbClient.GetDeviceById(token); err != nil {
			errorHandler.HandleManyVariants(
				w,
				err,
				[]errorconcept.ErrorConceptType{
					errorconcept.Database.NotFound,
					errorconcept.Database.InvalidObjectId,
				},
				errorconcept.Default.InternalServerError)
			return
		}
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(dev)
}

func decodeState(r *http.Request) (mode string, state string, err error) {
	var adminReq admin.UpdateRequest
	var opsReq operating.UpdateRequest

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	var errMsg string
	decoder := json.NewDecoder(bytes.NewBuffer(bodyBytes))
	err = decoder.Decode(&adminReq)
	if err != nil {
		switch err := err.(type) {
		case models.ErrContractInvalid:
			errMsg = err.Error()
		default:
			return "", "", err
		}
	} else {
		return ADMINSTATE, string(adminReq.AdminState), nil
	}

	// In this case, the supplied request was not for the AdminState. Try OperatingState.
	decoder = json.NewDecoder(bytes.NewBuffer(bodyBytes))
	err = decoder.Decode(&opsReq)
	if err != nil {
		switch err := err.(type) {
		case models.ErrContractInvalid:
			errMsg += "; " + err.Error()
		default:
			return "", "", err
		}
	} else {
		return OPSTATE, string(opsReq.OperatingState), nil
	}

	// In this case, the request we were given in completely invalid
	return "", "", fmt.Errorf(
		"unknown request type: data decode failed for both states: %v",
		errMsg)
}

func restSetDeviceStateById(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) {

	vars := mux.Vars(r)
	var did = vars[ID]
	updateMode, state, err := decodeState(r)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the device exists
	d, err := dbClient.GetDeviceById(did)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.ServiceUnavailable)
		return
	}

	if err = updateDeviceState(updateMode, state, d, dbClient); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.UpdateError_StatusInternalServer)
		return
	}

	// Notify
	_ = notifyDeviceAssociates(d, http.MethodPut, r.Context(), lc, dbClient, nc, configuration)

	w.WriteHeader(http.StatusOK)
}

func restSetDeviceStateByDeviceName(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) {

	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	updateMode, state, err := decodeState(r)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the device exists
	d, err := dbClient.GetDeviceByName(n)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.InternalServerError)
		return
	}

	if err = updateDeviceState(updateMode, state, d, dbClient); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.UpdateError_StatusInternalServer)
		return
	}

	ctx := r.Context()
	// Notify
	_ = notifyDeviceAssociates(d, http.MethodPut, ctx, lc, dbClient, nc, configuration)

	w.WriteHeader(http.StatusOK)
}

func updateDeviceState(updateMode string, state string, d models.Device, dbClient interfaces.DBClient) error {
	switch updateMode {
	case ADMINSTATE:
		d.AdminState = models.AdminState(strings.ToUpper(state))
	case OPSTATE:
		d.OperatingState = models.OperatingState(strings.ToUpper(state))
	}
	return dbClient.UpdateDevice(d)
}

func restDeleteDeviceById(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) {

	vars := mux.Vars(r)
	var did = vars[ID]

	// Check if the device exists
	d, err := dbClient.GetDeviceById(did)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.BadRequest)
		return
	}

	ctx := r.Context()
	if err := deleteDevice(d, w, ctx, lc, dbClient, errorHandler, nc, configuration); err != nil {
		lc.Error(err.Error())
		return
	}

	w.Write([]byte("true"))
}

func restDeleteDeviceByName(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) {

	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the device exists
	d, err := dbClient.GetDeviceByName(n)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Device.NotFound)
		return
	}

	ctx := r.Context()
	if err := deleteDevice(d, w, ctx, lc, dbClient, errorHandler, nc, configuration); err != nil {
		lc.Error(err.Error())
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Delete the device
func deleteDevice(
	d models.Device,
	w http.ResponseWriter,
	ctx context.Context,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) error {

	if err := deleteAssociatedReportsForDevice(d, w, lc, dbClient, errorHandler); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.DeleteError)
		return err
	}

	if err := dbClient.DeleteDeviceById(d.Id); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.DeleteError)
		return err
	}

	// Notify Associates
	err := notifyDeviceAssociates(d, http.MethodDelete, ctx, lc, dbClient, nc, configuration)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Device.NotifyError)
		return err
	}

	return nil
}

// Delete the associated device reports for the device
func deleteAssociatedReportsForDevice(
	d models.Device,
	w http.ResponseWriter,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) error {

	reports, err := dbClient.GetDeviceReportByDeviceName(d.Name)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusServiceUnavailable)
		return err
	}

	// Delete the associated reports
	for _, report := range reports {
		if err := dbClient.DeleteDeviceReportById(report.Id); err != nil {
			errorHandler.Handle(w, err, errorconcept.Common.DeleteError)
			return err
		}
		_ = notifyDeviceReportAssociates(report, http.MethodDelete, lc, dbClient)
	}

	return nil
}

func restSetDeviceLastConnectedById(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) {

	vars := mux.Vars(r)
	var did = vars[ID]
	var vlc = vars[LASTCONNECTED]
	lastConnected, err := strconv.ParseInt(vlc, 10, 64)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the device exists
	d, err := dbClient.GetDeviceById(did)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.InternalServerError)
		return
	}

	ctx := r.Context()
	// Update last connected
	err = setLastConnected(d, lastConnected, false, w, ctx, lc, dbClient, errorHandler, nc, configuration)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restSetLastConnectedByIdNotify(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) {

	// Get the URL parameters
	vars := mux.Vars(r)
	var did = vars[ID]
	var vlc = vars[LASTCONNECTED]
	notify, err := strconv.ParseBool(vars[LASTCONNECTEDNOTIFY])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}
	lastConnected, err := strconv.ParseInt(vlc, 10, 64)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the device exists
	d, err := dbClient.GetDeviceById(did)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.InternalServerError)
		return
	}

	ctx := r.Context()
	// Update last connected
	err = setLastConnected(d, lastConnected, notify, w, ctx, lc, dbClient, errorHandler, nc, configuration)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restSetDeviceLastConnectedByName(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) {

	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusServiceUnavailable)
		return
	}
	var vlc = vars[LASTCONNECTED]
	lastConnected, err := strconv.ParseInt(vlc, 10, 64)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	d, err := dbClient.GetDeviceByName(n)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.ServiceUnavailable)
		return
	}

	ctx := r.Context()
	// Update last connected
	err = setLastConnected(d, lastConnected, false, w, ctx, lc, dbClient, errorHandler, nc, configuration)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restSetDeviceLastConnectedByNameNotify(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) {

	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusServiceUnavailable)
		return
	}
	var vlc = vars[LASTCONNECTED]
	lastConnected, err := strconv.ParseInt(vlc, 10, 64)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusServiceUnavailable)
		return
	}
	notify, err := strconv.ParseBool(vars[LASTCONNECTEDNOTIFY])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	d, err := dbClient.GetDeviceByName(n)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.ServiceUnavailable)
		return
	}

	ctx := r.Context()
	// Update last connected
	err = setLastConnected(d, lastConnected, notify, w, ctx, lc, dbClient, errorHandler, nc, configuration)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the last connected value for the device
func setLastConnected(
	d models.Device,
	time int64,
	notify bool,
	w http.ResponseWriter,
	ctx context.Context,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) error {

	d.LastConnected = time
	if err := dbClient.UpdateDevice(d); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.UpdateError_StatusServiceUnavailable)
		return err
	}

	if notify {
		_ = notifyDeviceAssociates(d, http.MethodPut, ctx, lc, dbClient, nc, configuration)
	}

	return nil
}

func restSetDeviceLastReportedById(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) {

	vars := mux.Vars(r)
	var did = vars[ID]
	var vlr = vars[LASTREPORTED]
	lr, err := strconv.ParseInt(vlr, 10, 64)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	d, err := dbClient.GetDeviceById(did)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.ServiceUnavailable)
		return
	}

	ctx := r.Context()
	// Update Last Reported
	if err = setLastReported(
		d,
		lr,
		false,
		w,
		ctx,
		lc,
		dbClient,
		errorHandler,
		nc,
		configuration); err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restSetDeviceLastReportedByIdNotify(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) {

	vars := mux.Vars(r)
	var did = vars[ID]
	var vlr = vars[LASTREPORTED]
	lr, err := strconv.ParseInt(vlr, 10, 64)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusServiceUnavailable)
		return
	}
	notify, err := strconv.ParseBool(vars[LASTREPORTEDNOTIFY])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	d, err := dbClient.GetDeviceById(did)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.ServiceUnavailable)
		return
	}

	ctx := r.Context()
	// Update last reported
	err = setLastReported(d, lr, notify, w, ctx, lc, dbClient, errorHandler, nc, configuration)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restSetDeviceLastReportedByName(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) {

	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}
	var vlr = vars[LASTREPORTED]
	lr, err := strconv.ParseInt(vlr, 10, 64)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the device exists
	d, err := dbClient.GetDeviceByName(n)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.InternalServerError)
		return
	}

	ctx := r.Context()
	// Update last reported
	if err = setLastReported(
		d,
		lr,
		false,
		w,
		ctx,
		lc,
		dbClient,
		errorHandler,
		nc,
		configuration); err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restSetDeviceLastReportedByNameNotify(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) {

	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}
	var vlr = vars[LASTREPORTED]
	lr, err := strconv.ParseInt(vlr, 10, 64)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}
	notify, err := strconv.ParseBool(vars[LASTREPORTEDNOTIFY])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the device exists
	d, err := dbClient.GetDeviceByName(n)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.InternalServerError)
		return
	}

	ctx := r.Context()
	// Update last reported
	err = setLastReported(d, lr, notify, w, ctx, lc, dbClient, errorHandler, nc, configuration)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the last reported field of the device
func setLastReported(
	d models.Device,
	time int64,
	notify bool,
	w http.ResponseWriter,
	ctx context.Context,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) error {

	d.LastReported = time
	if err := dbClient.UpdateDevice(d); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.UpdateError_StatusServiceUnavailable)
		return err
	}

	if notify {
		_ = notifyDeviceAssociates(d, http.MethodPut, ctx, lc, dbClient, nc, configuration)
	}

	return nil
}

func restGetDeviceByName(
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
	res, err := dbClient.GetDeviceByName(dn)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.InternalServerError)
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(res)
}

// Notify the associated device service for the device
func notifyDeviceAssociates(
	d models.Device,
	action string,
	ctx context.Context,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) error {

	// Post the notification to the notifications service
	postNotification(d.Name, action, ctx, nc, configuration)

	// Callback for device service
	ds, err := dbClient.GetDeviceServiceById(d.Service.Id)
	if err != nil {
		lc.Error(err.Error())
		return err
	}
	if err := notifyAssociates([]models.DeviceService{ds}, d.Id, action, models.DEVICE, lc); err != nil {
		lc.Error(err.Error())
		return err
	}

	return nil
}

func postNotification(
	name string,
	action string,
	ctx context.Context,
	nc notifications.NotificationsClient,
	configuration *config.ConfigurationStruct) {

	// Only post notification if the configuration is set
	if configuration.Notifications.PostDeviceChanges {
		// Make the notification
		notification := notifications.Notification{
			Slug:        configuration.Notifications.Slug + strconv.FormatInt(db.MakeTimestamp(), 10),
			Content:     configuration.Notifications.Content + name + "-" + action,
			Category:    notifications.SW_HEALTH,
			Description: configuration.Notifications.Description,
			Labels:      []string{configuration.Notifications.Label},
			Sender:      configuration.Notifications.Sender,
			Severity:    notifications.NORMAL,
		}

		_ = nc.SendNotification(ctx, notification)
	}
}
