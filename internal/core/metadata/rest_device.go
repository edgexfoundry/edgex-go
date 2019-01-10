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
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	notifications "github.com/edgexfoundry/edgex-go/pkg/clients/notifications"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
)

func restGetAllDevices(w http.ResponseWriter, _ *http.Request) {
	res := make([]models.Device, 0)
	err := dbClient.GetAllDevices(&res)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check the max length
	if len(res) > Configuration.Service.ReadMaxLimit {
		err = errors.New("Max limit exceeded")
		LoggingClient.Error(err.Error())
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
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Addressable check
	// Try by name
	//TODO: Is what we really need here a checkForAddressable() function?
	addressable, err := dbClient.GetAddressableByName(d.Addressable.Name)
	if err != nil {
		// Try by ID
		addressable, err = dbClient.GetAddressableById(d.Addressable.Id)
		if err != nil {
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error()+": A device must be associated to an Addressable", http.StatusBadRequest)
			return
		}
	}
	d.Addressable = addressable

	// Service Check
	// Try by name
	service, err := dbClient.GetDeviceServiceByName(d.Service.Service.Name)
	if err != nil {
		// Try by ID
		service, err = dbClient.GetDeviceServiceById(d.Service.Service.Id)
		if err != nil {
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error()+": A device must be associated with a device service", http.StatusBadRequest)
			return
		}
	}
	d.Service = service

	// Profile Check
	// Try by name
	err = dbClient.GetDeviceProfileByName(&d.Profile, d.Profile.Name)
	if err != nil {
		// Try by ID
		err = dbClient.GetDeviceProfileById(&d.Profile, d.Profile.Id.Hex())
		if err != nil {
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error()+": A device must be associated with a device profile", http.StatusBadRequest)
			return
		}
	}

	// Check operating/admin state
	if d.OperatingState == models.OperatingState("") || d.AdminState == models.AdminState("") {
		err = errors.New("Device can't have null operating state or admin state")
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Add the device
	err = dbClient.AddDevice(&d)
	if err != nil {
		if err == db.ErrNotUnique {
			http.Error(w, "Duplicate name for device", http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
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
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var oldDevice models.Device
	// First try ID
	err = dbClient.GetDeviceById(&oldDevice, rd.Id.Hex())
	if err != nil {
		// Then try name
		err = dbClient.GetDeviceByName(&oldDevice, rd.Name)
		if err != nil {
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}

	if err = updateDeviceFields(rd, &oldDevice); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	if err = dbClient.UpdateDevice(oldDevice); err != nil {
		LoggingClient.Error(err.Error())
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
		a, err := dbClient.GetAddressableById(from.Addressable.Id)
		if err != nil {
			// Then try name
			a, err = dbClient.GetAddressableByName(from.Addressable.Name)
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
		ds, err := dbClient.GetDeviceServiceById(from.Service.Service.Id)
		if err != nil {
			// Then try name
			ds, err = dbClient.GetDeviceServiceByName(from.Service.Service.Name)
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
		err := dbClient.GetDeviceProfileById(&dp, from.Profile.Id.Hex())
		if err != nil {
			// Then try Name
			err = dbClient.GetDeviceProfileByName(&dp, from.Profile.Name)
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
		err := dbClient.GetDeviceByName(&checkD, from.Name)
		if err != nil {
			// A problem occurred accessing database
			if err != db.ErrNotFound {
				LoggingClient.Error(err.Error())
				return err
			}
		}

		// Found a device, make sure its the one we're trying to update
		if err != db.ErrNotFound {
			// Different IDs -> Name is not unique
			if checkD.Id != to.Id {
				err = errors.New("Duplicate name for Device")
				LoggingClient.Error(err.Error())
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
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := make([]models.Device, 0)
	err = dbClient.GetDevicesWithLabel(&res, label)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	err := dbClient.GetDeviceProfileById(&dp, pid)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	res := make([]models.Device, 0)
	err = dbClient.GetDevicesByProfileId(&res, pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
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
	_, err := dbClient.GetDeviceServiceById(sid)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		LoggingClient.Error(err.Error())
		return
	}

	err = dbClient.GetDevicesByServiceId(&res, sid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		LoggingClient.Error(err.Error())
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
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device service exists
	var ds models.DeviceService
	ds, err = dbClient.GetDeviceServiceByName(sn)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		LoggingClient.Error(err.Error())
		return
	}

	res := make([]models.Device, 0)

	// Find devices by service ID now that you have the Service object (and therefor the ID)
	err = dbClient.GetDevicesByServiceId(&res, ds.Service.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		LoggingClient.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetDeviceByAddressableName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	an, err := url.QueryUnescape(vars[ADDRESSABLENAME])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the addressable exists
	a, err := dbClient.GetAddressableByName(an)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		LoggingClient.Error(err.Error())
		return
	}

	res := make([]models.Device, 0)

	// Use the addressable ID now that you have the addressable object
	err = dbClient.GetDevicesByAddressableId(&res, a.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		LoggingClient.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetDeviceByProfileName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pn, err := url.QueryUnescape(vars[PROFILENAME])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device profile exists
	var dp models.DeviceProfile
	err = dbClient.GetDeviceProfileByName(&dp, pn)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		LoggingClient.Error(err.Error())
		return
	}

	res := make([]models.Device, 0)

	// Use profile ID now that you have the profile object
	err = dbClient.GetDevicesByProfileId(&res, dp.Id.Hex())
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		LoggingClient.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetDeviceByAddressableId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var aid string = vars[ADDRESSABLEID]

	// Check if the addressable exists
	_, err := dbClient.GetAddressableById(aid)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		LoggingClient.Error(err.Error())
		return
	}

	res := make([]models.Device, 0)
	err = dbClient.GetDevicesByAddressableId(&res, aid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		LoggingClient.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetDeviceById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars[ID]
	var res models.Device
	if err := dbClient.GetDeviceById(&res, did); err != nil {
		LoggingClient.Error(err.Error())
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

//Shouldn't need "rest" in any of these methods. Adding it here for consistency right now.
func restCheckForDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars[ID] //referring to this as "token" for now since the source variable is double purposed
	dev := models.Device{}
	//Check for name first since we're using that meaning by default.
	if err := dbClient.GetDeviceByName(&dev, token); err != nil {
		if err != db.ErrNotFound {
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			LoggingClient.Debug(fmt.Sprintf("device %s %v", token, err))
		}
	}
	//If lookup by name failed, see if we were passed the ID
	if len(dev.Name) == 0 {
		if bson.IsObjectIdHex(token) {
			if err := dbClient.GetDeviceById(&dev, token); err != nil {
				LoggingClient.Error(err.Error())
				if err == db.ErrNotFound {
					http.Error(w, err.Error(), http.StatusNotFound)
				} else {
					http.Error(w, err.Error(), http.StatusServiceUnavailable)
				}
				return
			}
		} else {
			http.Error(w, "device not found: "+token, http.StatusNotFound)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dev)
}

func restSetDeviceOpStateById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars[ID] // TODO check if DID needs to be a bson
	var os string = vars[OPSTATE]
	newOs, f := models.GetOperatingState(os)
	if !f {
		err := errors.New("Invalid State: " + os + " Must be 'ENABLED' or 'DISABLED'")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		LoggingClient.Error(err.Error())
		return
	}

	// Check if the device exists
	var d models.Device
	err := dbClient.GetDeviceById(&d, did)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		LoggingClient.Error(err.Error())
		return
	}

	// Update OpState
	d.OperatingState = newOs
	if err = dbClient.UpdateDevice(d); err != nil {
		return
	}
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Notify
	notifyDeviceAssociates(d, http.MethodPut)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
	return
}

func restSetDeviceOpStateByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var os string = vars[OPSTATE]
	newOs, f := models.GetOperatingState(os)
	// Opstate is invalid
	if !f {
		err := errors.New("Invalid State: " + os + " Must be 'ENABLED' or 'DISABLED'")
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the device exists
	var d models.Device
	err = dbClient.GetDeviceByName(&d, n)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	// Update OpState
	d.OperatingState = newOs
	if err = dbClient.UpdateDevice(d); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Notify
	notifyDeviceAssociates(d, http.MethodPut)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restSetDeviceAdminStateById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars[ID]
	var as string = vars[ADMINSTATE]
	newAs, f := models.GetAdminState(as)
	if !f {
		err := errors.New("Invalid State: " + as + " Must be 'LOCKED' or 'UNLOCKED'")
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var d models.Device
	err := dbClient.GetDeviceById(&d, did)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		LoggingClient.Error(err.Error())
		return
	}

	// Update the AdminState
	d.AdminState = newAs
	if err = dbClient.UpdateDevice(d); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	if err := notifyDeviceAssociates(d, http.MethodPut); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
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
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var as string = vars[ADMINSTATE]

	newAs, f := models.GetAdminState(as)
	if !f {
		err = errors.New("Invalid State: " + as + " Must be 'LOCKED' or 'UNLOCKED'")
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the device exists
	var d models.Device
	err = dbClient.GetDeviceByName(&d, n)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	d.AdminState = newAs
	// Update the admin state
	if err = dbClient.UpdateDevice(d); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := notifyDeviceAssociates(d, http.MethodPut); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
	return
}

func restDeleteDeviceById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars[ID]

	// Check if the device exists
	var d models.Device

	if err := dbClient.GetDeviceById(&d, did); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		LoggingClient.Error(err.Error())
		return
	}

	if err := deleteDevice(d, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}

	w.Write([]byte("true"))
}

func restDeleteDeviceByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the device exists
	var d models.Device
	err = dbClient.GetDeviceByName(&d, n)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := deleteDevice(d, w); err != nil {
		LoggingClient.Error(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Delete the device
func deleteDevice(d models.Device, w http.ResponseWriter) error {
	if err := deleteAssociatedReportsForDevice(d, w); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	if err := dbClient.DeleteDeviceById(d.Id.Hex()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	// Notify Associates
	if err := notifyDeviceAssociates(d, http.MethodDelete); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	return nil
}

// Delete the associated device reports for the device
func deleteAssociatedReportsForDevice(d models.Device, w http.ResponseWriter) error {
	reports, err := dbClient.GetDeviceReportByDeviceName(d.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		LoggingClient.Error(err.Error())
		return err
	}

	// Delete the associated reports
	for _, report := range reports {
		if err := dbClient.DeleteDeviceReportById(report.Id); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			LoggingClient.Error(err.Error())
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
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the device exists
	var d models.Device
	err = dbClient.GetDeviceById(&d, did)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
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
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	lc, err := strconv.ParseInt(vlc, 10, 64)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the device exists
	var d models.Device
	err = dbClient.GetDeviceById(&d, did)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
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
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	var vlc string = vars[LASTCONNECTED]
	lc, err := strconv.ParseInt(vlc, 10, 64)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var d models.Device
	err = dbClient.GetDeviceByName(&d, n)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		LoggingClient.Error(err.Error())
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
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	var vlc string = vars[LASTCONNECTED]
	lc, err := strconv.ParseInt(vlc, 10, 64)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	notify, err := strconv.ParseBool(vars[LASTCONNECTEDNOTIFY])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var d models.Device
	err = dbClient.GetDeviceByName(&d, n)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		LoggingClient.Error(err.Error())
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
	d.LastConnected = time
	if err := dbClient.UpdateDevice(d); err != nil {
		LoggingClient.Error(err.Error())
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
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var d models.Device
	err = dbClient.GetDeviceById(&d, did)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		LoggingClient.Error(err.Error())
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
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	notify, err := strconv.ParseBool(vars[LASTREPORTEDNOTIFY])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the device exists
	var d models.Device
	err = dbClient.GetDeviceById(&d, did)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		LoggingClient.Error(err.Error())
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
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var vlr string = vars[LASTREPORTED]
	lr, err := strconv.ParseInt(vlr, 10, 64)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the device exists
	var d models.Device
	err = dbClient.GetDeviceByName(&d, n)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
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
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var vlr string = vars[LASTREPORTED]
	lr, err := strconv.ParseInt(vlr, 10, 64)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	notify, err := strconv.ParseBool(vars[LASTREPORTEDNOTIFY])
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the device exists
	var d models.Device
	err = dbClient.GetDeviceByName(&d, n)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
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
	d.LastReported = time
	if err := dbClient.UpdateDevice(d); err != nil {
		LoggingClient.Error(err.Error())
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
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var res models.Device
	err = dbClient.GetDeviceByName(&res, dn)
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

// Notify the associated device service for the device
func notifyDeviceAssociates(d models.Device, action string) error {
	// Post the notification to the notifications service
	postNotification(d.Name, action)

	// Callback for device service
	var ds models.DeviceService
	var err error
	if ds, err = dbClient.GetDeviceServiceById(d.Service.Service.Id); err != nil {
		LoggingClient.Error(err.Error())
		return err
	}
	var services []models.DeviceService
	services = append(services, ds)
	if err := notifyAssociates(services, d.Id.Hex(), action, models.DEVICE); err != nil {
		LoggingClient.Error(err.Error())
		return err
	}

	return nil
}

func postNotification(name string, action string) {
	// Only post notification if the configuration is set
	if Configuration.Notifications.PostDeviceChanges {
		// Make the notification
		notification := notifications.Notification{
			Slug:        Configuration.Notifications.Slug + strconv.FormatInt(db.MakeTimestamp(), 10),
			Content:     Configuration.Notifications.Content + name + "-" + string(action),
			Category:    notifications.SW_HEALTH,
			Description: Configuration.Notifications.Description,
			Labels:      []string{Configuration.Notifications.Label},
			Sender:      Configuration.Notifications.Sender,
			Severity:    notifications.NORMAL,
		}

		nc.SendNotification(notification)
	}
}
