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

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
)

func restGetAllDeviceReports(w http.ResponseWriter, _ *http.Request) {
	res, err := dbClient.GetAllDeviceReports()
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check max limit
	if len(res) > Configuration.Service.ReadMaxLimit {
		err = errors.New("Max limit exceeded")
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		LoggingClient.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&res)
}

// Add a new device report
// Referenced objects (Device, Schedule event) must already exist
// 404 If any of the referenced objects aren't found
func restAddDeviceReport(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var dr models.DeviceReport
	if err := json.NewDecoder(r.Body).Decode(&dr); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the device exists
	var d models.Device
	if err := dbClient.GetDeviceByName(&d, dr.Device); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Device referenced by Device Report doesn't exist", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	// Check if the Schedule Event exists
	var se models.ScheduleEvent
	if err := dbClient.GetScheduleEventByName(&se, dr.Event); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Schedule Event referenced by Device Report doesn't exist", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	// Add the device report
	var err error
	dr.Id, err = dbClient.AddDeviceReport(dr)
	if err != nil {
		if err == db.ErrNotUnique {
			http.Error(w, "Duplicate Name for the device report", http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	// Notify associates
	if err := notifyDeviceReportAssociates(dr, http.MethodPost); err != nil {
		LoggingClient.Error(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(dr.Id))
}

func restUpdateDeviceReport(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var from models.DeviceReport
	if err := json.NewDecoder(r.Body).Decode(&from); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the device report exists
	// First try ID
	var to models.DeviceReport
	to, err := dbClient.GetDeviceReportById(from.Id)
	if err != nil {
		// Try by name
		to, err = dbClient.GetDeviceReportByName(from.Name)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}
	}

	if err := updateDeviceReportFields(from, &to, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}

	if err := dbClient.UpdateDeviceReport(to); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Notify Associates
	if err := notifyDeviceReportAssociates(to, http.MethodPut); err != nil {
		LoggingClient.Error(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the relevant fields for the device report
func updateDeviceReportFields(from models.DeviceReport, to *models.DeviceReport, w http.ResponseWriter) error {
	if from.Device != "" {
		to.Device = from.Device
		if err := validateDevice(to.Device, w); err != nil {
			return err
		}
	}
	if from.Event != "" {
		to.Event = from.Event
		if err := validateEvent(to.Event, w); err != nil {
			return err
		}
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
func validateDevice(d string, w http.ResponseWriter) error {
	var device models.Device
	if err := dbClient.GetDeviceByName(&device, d); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Device was not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return err
	}

	return nil
}

// Validate that the schedule event exists
func validateEvent(e string, w http.ResponseWriter) error {
	var event models.ScheduleEvent
	if err := dbClient.GetScheduleEventByName(&event, e); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Event was not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return err
	}

	return nil
}

func restGetReportById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars[ID]
	res, err := dbClient.GetDeviceReportById(did)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetReportByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	res, err := dbClient.GetDeviceReportByName(n)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Get a list of value descriptor names
// The names are a union of all the value descriptors from the device reports for the given device
func restGetValueDescriptorsForDeviceName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[DEVICENAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	// Get all the associated device reports
	reports, err := dbClient.GetDeviceReportByDeviceName(n)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	valueDescriptors := []string{}
	for _, report := range reports {
		for _, e := range report.Expected {
			valueDescriptors = append(valueDescriptors, e)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(valueDescriptors)
}
func restGetDeviceReportByDeviceName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[DEVICENAME])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := dbClient.GetDeviceReportByDeviceName(n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restDeleteReportById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var id string = vars[ID]

	// Check if the device report exists
	dr, err := dbClient.GetDeviceReportById(id)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	if err := deleteDeviceReport(dr, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restDeleteReportByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the device report exists
	dr, err := dbClient.GetDeviceReportByName(n)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	if err = deleteDeviceReport(dr, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func deleteDeviceReport(dr models.DeviceReport, w http.ResponseWriter) error {
	if err := dbClient.DeleteDeviceReportById(dr.Id); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	// Notify Associates
	if err := notifyDeviceReportAssociates(dr, http.MethodDelete); err != nil {
		return err
	}

	return nil
}

// Notify the associated device services to the device report
func notifyDeviceReportAssociates(dr models.DeviceReport, action string) error {
	// Get the device of the report
	var d models.Device
	var err error
	if err = dbClient.GetDeviceByName(&d, dr.Device); err != nil {
		return err
	}

	// Get the device service for the device
	var ds models.DeviceService
	if ds, err = dbClient.GetDeviceServiceById(d.Service.Service.Id); err != nil {
		return err
	}

	var services []models.DeviceService
	services = append(services, ds)

	// Notify the associating device services
	if err = notifyAssociates(services, dr.Id, action, models.REPORT); err != nil {
		return err
	}

	return nil
}
