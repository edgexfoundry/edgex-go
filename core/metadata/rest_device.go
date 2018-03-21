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
 *
 * @microservice: core-metadata-go service
 * @author: Spencer Bull & Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package metadata

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
	notifications "github.com/edgexfoundry/edgex-go/support/notifications-client"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
)

func restGetAllDevices(w http.ResponseWriter, _ *http.Request) {
	res := make([]models.Device, 0)
	err := getAllDevices(&res)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check the max length
	if len(res) > configuration.ReadMaxLimit {
		err = errors.New("Max limit exceeded")
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&res)
}

// Post a new device
// Attached objects (Addressable, Profile, Service) are referenced by ID or name
// 409 conflict if any of the attached items can't be found by ID or name
// Ignore everything else from the attached objects
func restAddNewDevice(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var d models.Device
	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Addressable check
	// Try by name
	err = getAddressableByName(&d.Addressable, d.Addressable.Name)
	if err != nil {
		// Try by ID
		err = getAddressableById(&d.Addressable, d.Addressable.Id.Hex())
		if err != nil {
			loggingClient.Error(err.Error(), "")
			http.Error(w, err.Error()+": A device must be associated to an Addressable", http.StatusConflict)
			return
		}
	}

	// Service Check
	// Try by name
	err = getDeviceServiceByName(&d.Service, d.Service.Service.Name)
	if err != nil {
		// Try by ID
		err = getDeviceServiceById(&d.Service, d.Service.Service.Id.Hex())
		if err != nil {
			loggingClient.Error(err.Error(), "")
			http.Error(w, err.Error()+": A device must be associated with a device service", http.StatusConflict)
			return
		}
	}

	// Profile Check
	// Try by name
	err = getDeviceProfileByName(&d.Profile, d.Profile.Name)
	if err != nil {
		// Try by ID
		err = getDeviceProfileById(&d.Profile, d.Profile.Id.Hex())
		if err != nil {
			loggingClient.Error(err.Error(), "")
			http.Error(w, err.Error()+": A device must be associated with a device profile", http.StatusConflict)
			return
		}
	}

	// Check operating/admin state
	if d.OperatingState == models.OperatingState("") || d.AdminState == models.AdminState("") {
		err = errors.New("Device can't have null operating state or admin state")
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// Add the device
	err = addDevice(&d)
	if err != nil {
		if err == ErrDuplicateName {
			http.Error(w, "Duplicate name for device", http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	// Notify the associates
	notifyDeviceAssociates(d, http.MethodPost)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(d.Id.Hex()))
}

// Update the device
// Use ID to identify device first, then name
// Can't create new Device Services/Profiles with a PUT, but you can reference another one
func restUpdateDevice(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var rd models.Device
	err := json.NewDecoder(r.Body).Decode(&rd)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var oldDevice models.Device
	// First try ID
	err = getDeviceById(&oldDevice, rd.Id.Hex())
	if err != nil {
		// Then try name
		err = getDeviceByName(&oldDevice, rd.Name)
		if err != nil {
			loggingClient.Error(err.Error(), "")
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}

	if err = updateDeviceFields(rd, &oldDevice); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	if err = UpdateDevice(oldDevice); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Notify
	notifyDeviceAssociates(oldDevice, http.MethodPut)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the device fields
func updateDeviceFields(from models.Device, to *models.Device) error {
	if (from.Addressable != models.Addressable{}) {
		// Check if the new addressable exists
		var a models.Addressable
		// Try ID first
		err := getAddressableById(&a, from.Addressable.Id.Hex())
		if err != nil {
			// Then try name
			err = getAddressableByName(&a, from.Addressable.Name)
			if err != nil {
				return errors.New("Addressable not found for updated device")
			}
		}

		to.Addressable = a
	}
	if (from.Service.String() != models.DeviceService{}.String()) {
		// Check if the new service exists
		var ds models.DeviceService
		// Try ID first
		err := getDeviceServiceById(&ds, from.Service.Service.Id.Hex())
		if err != nil {
			// Then try name
			err = getDeviceServiceByName(&ds, from.Service.Service.Name)
			if err != nil {
				return errors.New("Device service not found for updated device")
			}
		}

		to.Service = ds
	}
	if (from.Profile.String() != models.DeviceProfile{}.String()) {
		// Check if the new profile exists
		var dp models.DeviceProfile
		// Try ID first
		err := getDeviceProfileById(&dp, from.Profile.Id.Hex())
		if err != nil {
			// Then try Name
			err = getDeviceProfileByName(&dp, from.Profile.Name)
			if err != nil {
				return errors.New("Device profile not found for updated device")
			}
		}

		to.Profile = dp
	}
	if from.AdminState != "" {
		to.AdminState = from.AdminState
	}
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
	if from.Location != nil {
		to.Location = from.Location
	}
	if from.OperatingState != models.OperatingState("") {
		to.OperatingState = from.OperatingState
	}
	if from.Origin != 0 {
		to.Origin = from.Origin
	}
	if from.Name != "" {
		to.Name = from.Name

		// Check if the name is unique
		var checkD models.Device
		err := getDeviceByName(&checkD, from.Name)
		if err != nil {
			// A problem occured accessing database
			if err != mgo.ErrNotFound {
				loggingClient.Error(err.Error(), "")
				return err
			}
		}

		// Found a device, make sure its the one we're trying to update
		if err != mgo.ErrNotFound {
			// Differnt IDs -> Name is not unique
			if checkD.Id != to.Id {
				err = errors.New("Duplicate name for Device")
				loggingClient.Error(err.Error(), "")
				return err
			}
		}
	}

	return nil
}

func restGetDevicesWithLabel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	label, err := url.QueryUnescape(vars[LABEL])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	var labels []string
	labels = append(labels, label)

	res := make([]models.Device, 0)
	err = getDevicesWithLabel(&res, labels)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetDeviceByProfileId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var pid string = vars[PROFILEID]

	// Check if the device profile exists
	var dp models.DeviceProfile
	err := getDeviceProfileById(&dp, pid)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	res := make([]models.Device, 0)
	err = getDevicesByProfileId(&res, pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetDeviceByServiceId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var sid string = vars[SERVICEID]
	res := make([]models.Device, 0)

	// Check if the device service exists
	var ds models.DeviceService
	err := getDeviceServiceById(&ds, sid)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	err = getDevicesByServiceId(&res, sid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// If the result array is empty, don't return http.NotFound, just return empty array
func restGetDeviceByServiceName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sn, err := url.QueryUnescape(vars[SERVICENAME])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device service exists
	var ds models.DeviceService
	err = getDeviceServiceByName(&ds, sn)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	res := make([]models.Device, 0)

	// Find devices by service ID now that you have the Service object (and therefor the ID)
	err = getDevicesByServiceId(&res, ds.Service.Id.Hex())
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetDeviceByAddressableName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	an, err := url.QueryUnescape(vars[ADDRESSABLENAME])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the addressable exists
	var a models.Addressable
	err = getAddressableByName(&a, an)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	res := make([]models.Device, 0)

	// Use the addressable ID now that you have the addressable object
	err = getDevicesByAddressableId(&res, a.Id.Hex())
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetDeviceByProfileName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pn, err := url.QueryUnescape(vars[PROFILENAME])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device profile exists
	var dp models.DeviceProfile
	err = getDeviceProfileByName(&dp, pn)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	res := make([]models.Device, 0)

	// Use profile ID now that you have the profile object
	err = getDevicesByProfileId(&res, dp.Id.Hex())
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetDeviceByAddressableId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var aid string = vars[ADDRESSABLEID]

	// Check if the addressable exists
	var a models.Addressable
	err := getAddressableById(&a, aid)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	res := make([]models.Device, 0)
	err = getDevicesByAddressableId(&res, aid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetDeviceById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars[ID]
	var res models.Device
	if err := getDeviceById(&res, did); err != nil {
		loggingClient.Error(err.Error(), "")
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restSetDeviceOpStateById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars[ID] // TODO check if DID needs to be a bson
	var os string = vars[OPSTATE]
	f := models.IsOperatingStateType(os)
	if !f {
		err := errors.New("Invalid State: " + os + " Must be 'ENABLED' or 'DISABLED'")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	// Check if the device exists
	var d models.Device
	err := getDeviceById(&d, did)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	// Update OpState
	if err = setOpState(d, os, w); err != nil {
		return
	}
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
	return
}

func restSetDeviceOpStateByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	var os string = vars[OPSTATE]
	f := models.IsOperatingStateType(os)
	// Opstate is invalid
	if !f {
		err := errors.New("Invalid State: " + os + " Must be 'ENABLED' or 'DISABLED'")
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var d models.Device
	err = getDeviceByName(&d, n)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	// Update OpState
	if err = setOpState(d, os, w); err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the opstate of the device
func setOpState(d models.Device, os string, w http.ResponseWriter) error {
	err := setById(DEVICECOL, d.Id.Hex(), "operatingState", os)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	// Notify
	notifyDeviceAssociates(d, http.MethodPut)

	return nil
}

func restSetDeviceAdminStateById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars[ID]
	var as string = vars[ADMINSTATE]
	f := models.IsAdminStateType(as)
	if !f {
		err := errors.New("Invalid State: " + as + " Must be 'LOCKED' or 'UNLOCKED'")
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var d models.Device
	err := getDeviceById(&d, did)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	// Update the AdminState
	if err = setAdminState(d, as, w); err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
	return
}

func restSetDeviceAdminStateByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	var as string = vars[ADMINSTATE]

	f := models.IsAdminStateType(as)
	if !f {
		err = errors.New("Invalid State: " + as + " Must be 'LOCKED' or 'UNLOCKED'")
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var d models.Device
	err = getDeviceByName(&d, n)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	// Update the admin state
	if err = setAdminState(d, as, w); err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
	return
}

// Update the admin state for the device
func setAdminState(d models.Device, as string, w http.ResponseWriter) error {
	if err := setById(DEVICECOL, d.Id.Hex(), ADMINSTATE, as); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	if err := notifyDeviceAssociates(d, http.MethodPut); err != nil {
		return err
	}

	return nil
}

func restDeleteDeviceById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars[ID]

	// Check if the device exists
	var d models.Device
	//err := getDeviceById(&d, did)
	if err := getDeviceById(&d, did); err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	if err := deleteDevice(d, w); err != nil {
		loggingClient.Error(err.Error(), "")
		return
	}

	w.Write([]byte("true"))
}

func restDeleteDeviceByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var d models.Device
	err = getDeviceByName(&d, n)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := deleteDevice(d, w); err != nil {
		loggingClient.Error(err.Error(), "")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Delete the device
func deleteDevice(d models.Device, w http.ResponseWriter) error {
	if err := deleteAssociatedReportsForDevice(d, w); err != nil {
		return err
	}

	if err := deleteById(DEVICECOL, d.Id.Hex()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	// Notify Associates
	if err := notifyDeviceAssociates(d, http.MethodDelete); err != nil {
		return err
	}

	return nil
}

// Delete the associated device reports for the device
func deleteAssociatedReportsForDevice(d models.Device, w http.ResponseWriter) error {
	var reports []models.DeviceReport
	if err := getDeviceReportByDeviceName(&reports, d.Name); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return err
	}

	// Delete the associated reports
	for _, report := range reports {
		if err := deleteById(DRCOL, report.Id.Hex()); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error(), "")
			return err
		}
		notifyDeviceReportAssociates(report, http.MethodDelete)
	}

	return nil
}

func restSetDeviceLastConnectedById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars[ID]
	var vlc string = vars[LASTCONNECTED]
	lc, err := strconv.ParseInt(vlc, 10, 64)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var d models.Device
	err = getDeviceById(&d, did)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	// Update last connected
	if err = setLastConnected(d, lc, false, w); err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restSetLastConnectedByIdNotify(w http.ResponseWriter, r *http.Request) {
	// Get the URL parameters
	vars := mux.Vars(r)
	var did = vars[ID]
	var vlc = vars[LASTCONNECTED]
	notify, err := strconv.ParseBool(vars[LASTCONNECTEDNOTIFY])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	lc, err := strconv.ParseInt(vlc, 10, 64)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var d models.Device
	err = getDeviceById(&d, did)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	// Update last connected
	if err = setLastConnected(d, lc, notify, w); err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restSetDeviceLastConnectedByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	var vlc string = vars[LASTCONNECTED]
	lc, err := strconv.ParseInt(vlc, 10, 64)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var d models.Device
	err = getDeviceByName(&d, n)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	// Update last connected
	if err = setLastConnected(d, lc, false, w); err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restSetDeviceLastConnectedByNameNotify(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	var vlc string = vars[LASTCONNECTED]
	lc, err := strconv.ParseInt(vlc, 10, 64)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	notify, err := strconv.ParseBool(vars[LASTCONNECTEDNOTIFY])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var d models.Device
	err = getDeviceByName(&d, n)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	// Update last connected
	if err = setLastConnected(d, lc, notify, w); err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the last connected value for the device
func setLastConnected(d models.Device, time int64, notify bool, w http.ResponseWriter) error {
	if err := setByIdInt(DEVICECOL, d.Id.Hex(), LASTCONNECTED, time); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	if notify {
		notifyDeviceAssociates(d, http.MethodPut)
	}

	return nil
}

func restSetDeviceLastReportedById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars[ID]
	var vlr string = vars[LASTREPORTED]
	lr, err := strconv.ParseInt(vlr, 10, 64)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var d models.Device
	err = getDeviceById(&d, did)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	// Update Last Reported
	if err = setLastReported(d, lr, false, w); err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restSetDeviceLastReportedByIdNotify(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars[ID]
	var vlr string = vars[LASTREPORTED]
	lr, err := strconv.ParseInt(vlr, 10, 64)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	notify, err := strconv.ParseBool(vars[LASTREPORTEDNOTIFY])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var d models.Device
	err = getDeviceById(&d, did)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	// Update last reported
	if err = setLastReported(d, lr, notify, w); err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restSetDeviceLastReportedByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	var vlr string = vars[LASTREPORTED]
	lr, err := strconv.ParseInt(vlr, 10, 64)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var d models.Device
	err = getDeviceByName(&d, n)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	// Update last reported
	if err = setLastReported(d, lr, false, w); err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restSetDeviceLastReportedByNameNotify(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	var vlr string = vars[LASTREPORTED]
	lr, err := strconv.ParseInt(vlr, 10, 64)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	notify, err := strconv.ParseBool(vars[LASTREPORTEDNOTIFY])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var d models.Device
	err = getDeviceByName(&d, n)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}

	// Update last reported
	if err = setLastReported(d, lr, notify, w); err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the last reported field of the device
func setLastReported(d models.Device, time int64, notify bool, w http.ResponseWriter) error {
	if err := setByIdInt(DEVICECOL, d.Id.Hex(), LASTREPORTED, time); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	if notify {
		notifyDeviceAssociates(d, http.MethodPut)
	}

	return nil
}

func restGetDeviceByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dn, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	var res models.Device
	err = getDeviceByName(&res, dn)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		loggingClient.Error(err.Error(), "")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Notify the associated device service for the device
func notifyDeviceAssociates(d models.Device, action string) error {
	// Post the notification to the notifications service
	postNotification(d.Name, action)

	// Callback for device service
	var ds models.DeviceService
	if err := getDeviceServiceById(&ds, d.Service.Service.Id.Hex()); err != nil {
		loggingClient.Error(err.Error(), "")
		return err
	}
	var services []models.DeviceService
	services = append(services, ds)
	if err := notifyAssociates(services, d.Id.Hex(), action, models.DEVICE); err != nil {
		loggingClient.Error(err.Error(), "")
		return err
	}

	return nil
}

func postNotification(name string, action string) {
	// Only post notification if the configuration is set
	if configuration.NotificationPostDeviceChanges {
		// Make the notification
		notification := notifications.Notification{
			Slug:        configuration.NotificationsSlug + strconv.FormatInt((time.Now().UnixNano()/int64(time.Millisecond)), 10),
			Content:     configuration.NotificationContent + name + "-" + string(action),
			Category:    notifications.SW_HEALTH,
			Description: configuration.NotificationDescription,
			Labels:      []string{configuration.NotificationLabel},
			Sender:      configuration.NotificationSender,
			Severity:    notifications.NORMAL,
		}

		notificationsClient.RecieveNotification(notification)
	}
}
