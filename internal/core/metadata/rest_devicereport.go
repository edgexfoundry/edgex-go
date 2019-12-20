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
	"net/http"
	"net/url"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/config"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/gorilla/mux"
)

func restGetAllDeviceReports(
	w http.ResponseWriter,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler,
	configuration *config.ConfigurationStruct) {

	res, err := dbClient.GetAllDeviceReports()
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusInternalServer)
		return
	}

	// Check max limit
	if err = checkMaxLimit(len(res), configuration); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.LimitExceeded)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(&res)
}

// Add a new device report
// Referenced objects (Device, Schedule event) must already exist
// 404 If any of the referenced objects aren't found
func restAddDeviceReport(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	defer r.Body.Close()
	var dr models.DeviceReport
	if err := json.NewDecoder(r.Body).Decode(&dr); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the device exists
	if _, err := dbClient.GetDeviceByName(dr.Device); err != nil {
		errorHandler.HandleOneVariant(
			w,
			err,
			errorconcept.DeviceReport.DeviceNotFound,
			errorconcept.Default.InternalServerError)
		return
	}

	// Add the device report
	var err error
	dr.Id, err = dbClient.AddDeviceReport(dr)
	if err != nil {
		errorHandler.HandleOneVariant(
			w,
			err,
			errorconcept.DeviceReport.NotUnique,
			errorconcept.Default.InternalServerError)
		return
	}

	// Notify associates
	if err := notifyDeviceReportAssociates(dr, http.MethodPost, lc, dbClient); err != nil {
		lc.Error(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(dr.Id))
}

func restUpdateDeviceReport(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	defer r.Body.Close()
	var from models.DeviceReport
	if err := json.NewDecoder(r.Body).Decode(&from); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Check if the device report exists
	// First try ID
	to, err := dbClient.GetDeviceReportById(from.Id)
	if err != nil {
		// Try by name
		to, err = dbClient.GetDeviceReportByName(from.Name)
		if err != nil {
			errorHandler.HandleOneVariant(
				w,
				err,
				errorconcept.Database.NotFound,
				errorconcept.Default.InternalServerError)
			return
		}
	}

	if err := updateDeviceReportFields(from, &to, w, dbClient, errorHandler); err != nil {
		lc.Error(err.Error())
		return
	}

	if err := dbClient.UpdateDeviceReport(to); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.UpdateError_StatusInternalServer)
		return
	}

	// Notify Associates
	if err := notifyDeviceReportAssociates(to, http.MethodPut, lc, dbClient); err != nil {
		lc.Error(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the relevant fields for the device report
func updateDeviceReportFields(
	from models.DeviceReport,
	to *models.DeviceReport,
	w http.ResponseWriter,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) error {

	if from.Device != "" {
		to.Device = from.Device
		if err := validateDevice(to.Device, w, dbClient, errorHandler); err != nil {
			return err
		}
	}
	if from.Action != "" {
		to.Action = from.Action
	}
	if from.Expected != nil {
		to.Expected = from.Expected
		// TODO: Someday find a way to check the value descriptors
	}
	if from.Name != "" {
		to.Name = from.Name
	}
	if from.Origin != 0 {
		to.Origin = from.Origin
	}

	return nil
}

// Validate that the device exists
func validateDevice(
	d string,
	w http.ResponseWriter,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) error {

	if _, err := dbClient.GetDeviceByName(d); err != nil {
		errorHandler.HandleOneVariant(
			w,
			err,
			errorconcept.DeviceReport.DeviceNotFound,
			errorconcept.Default.ServiceUnavailable)
		return err
	}

	return nil
}

func restGetReportById(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	var did string = vars[ID]
	res, err := dbClient.GetDeviceReportById(did)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.InternalServerError)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(res)
}

func restGetReportByName(
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

	res, err := dbClient.GetDeviceReportByName(n)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.InternalServerError)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(res)
}

// Get a list of value descriptor names
// The names are a union of all the value descriptors from the device reports for the given device
func restGetValueDescriptorsForDeviceName(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[DEVICENAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	// Get all the associated device reports
	reports, err := dbClient.GetDeviceReportByDeviceName(n)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusInternalServer)
		return
	}

	valueDescriptors := []string{}
	for _, report := range reports {
		for _, e := range report.Expected {
			valueDescriptors = append(valueDescriptors, e)
		}
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(valueDescriptors)
}
func restGetDeviceReportByDeviceName(
	w http.ResponseWriter,
	r *http.Request,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[DEVICENAME])
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.InvalidRequest_StatusBadRequest)
		return
	}

	res, err := dbClient.GetDeviceReportByDeviceName(n)
	if err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.RetrieveError_StatusInternalServer)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	json.NewEncoder(w).Encode(res)
}

func restDeleteReportById(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) {

	vars := mux.Vars(r)
	var id string = vars[ID]

	// Check if the device report exists
	dr, err := dbClient.GetDeviceReportById(id)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.InternalServerError)
		return
	}

	if err := deleteDeviceReport(dr, w, lc, dbClient, errorHandler); err != nil {
		lc.Error(err.Error())
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restDeleteReportByName(
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

	// Check if the device report exists
	dr, err := dbClient.GetDeviceReportByName(n)
	if err != nil {
		errorHandler.HandleOneVariant(w, err, errorconcept.Database.NotFound, errorconcept.Default.InternalServerError)
		return
	}

	if err = deleteDeviceReport(dr, w, lc, dbClient, errorHandler); err != nil {
		lc.Error(err.Error())
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func deleteDeviceReport(
	dr models.DeviceReport,
	w http.ResponseWriter,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	errorHandler errorconcept.ErrorHandler) error {

	if err := dbClient.DeleteDeviceReportById(dr.Id); err != nil {
		errorHandler.Handle(w, err, errorconcept.Common.DeleteError)
		return err
	}

	// Notify Associates
	if err := notifyDeviceReportAssociates(dr, http.MethodDelete, lc, dbClient); err != nil {
		return err
	}

	return nil
}

// Notify the associated device services to the device report
func notifyDeviceReportAssociates(
	dr models.DeviceReport,
	action string,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) error {

	// Get the device of the report
	d, err := dbClient.GetDeviceByName(dr.Device)
	if err != nil {
		return err
	}

	// Get the device service for the device
	ds, err := dbClient.GetDeviceServiceById(d.Service.Id)
	if err != nil {
		return err
	}

	// Notify the associating device services
	if err = notifyAssociates([]models.DeviceService{ds}, dr.Id, action, models.REPORT, lc); err != nil {
		return err
	}

	return nil
}
