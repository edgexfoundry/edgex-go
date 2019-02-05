//
// Copyright (c) 2017
// Mainflux
// Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
)

const (
	distroPort int = 48070
)

const (
	typeAlgorithms   = "algorithms"
	typeCompressions = "compressions"
	typeFormats      = "formats"
	typeDestinations = "destinations"

	applicationJson = "application/json; charset=utf-8"
)

func getRegByID(w http.ResponseWriter, r *http.Request) {
	// URL parameters
	vars := mux.Vars(r)
	id := vars["id"]

	reg, err := dbClient.RegistrationById(id)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query by id: %s. Error: %s", id, err.Error()))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", applicationJson)
	json.NewEncoder(w).Encode(&reg)
}

func getRegList(w http.ResponseWriter, r *http.Request) {
	// URL parameters
	vars := mux.Vars(r)
	t := vars["type"]

	var list []string

	switch t {
	case typeAlgorithms:
		list = append(list, models.EncNone)
		list = append(list, models.EncAes)
	case typeCompressions:
		list = append(list, models.CompNone)
		list = append(list, models.CompGzip)
		list = append(list, models.CompZip)
	case typeFormats:
		list = append(list, models.FormatJSON)
		list = append(list, models.FormatXML)
		list = append(list, models.FormatIoTCoreJSON)
		list = append(list, models.FormatAzureJSON)
		list = append(list, models.FormatAWSJSON)
		list = append(list, models.FormatThingsBoardJSON)
		list = append(list, models.FormatNOOP)
	case typeDestinations:
		list = append(list, models.DestMQTT)
		list = append(list, models.DestIotCoreMQTT)
		list = append(list, models.DestAzureMQTT)
		list = append(list, models.DestRest)
		list = append(list, models.DestXMPP)
		list = append(list, models.DestAWSMQTT)
	default:
		LoggingClient.Error("Unknown type: " + t)
		http.Error(w, "Unknown type: "+t, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", applicationJson)
	json.NewEncoder(w).Encode(&list)
}

func getAllReg(w http.ResponseWriter, r *http.Request) {
	reg, err := dbClient.Registrations()
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query all registrations. Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", applicationJson)
	json.NewEncoder(w).Encode(&reg)
}

func getRegByName(w http.ResponseWriter, r *http.Request) {
	// URL parameters
	vars := mux.Vars(r)
	name := vars["name"]

	reg, err := dbClient.RegistrationByName(name)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query by name. Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", applicationJson)
	json.NewEncoder(w).Encode(&reg)
}

func addReg(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query add registration. Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reg := models.Registration{}
	if err := json.Unmarshal(data, &reg); err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query add registration. Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if valid, err := reg.Validate(); !valid {
		LoggingClient.Error(fmt.Sprintf("Failed to validate registrations fields: %X. Error: %s", data, err.Error()))
		http.Error(w, "Could not validate json fields", http.StatusBadRequest)
		return
	}

	_, err = dbClient.RegistrationByName(reg.Name)
	if err == nil {
		LoggingClient.Error("Name already taken: " + reg.Name)
		http.Error(w, "Name already taken", http.StatusBadRequest)
		return
	} else if err != db.ErrNotFound {
		LoggingClient.Error(fmt.Sprintf("Failed to query add registration. Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := dbClient.AddRegistration(reg)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query add registration. Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	notifyUpdatedRegistrations(models.NotifyUpdate{Name: reg.Name,
		Operation: "add"})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(id))
}

func updateReg(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to read update registration. Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var fromReg models.Registration
	if err := json.Unmarshal(data, &fromReg); err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to unmarshal update registration. Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the registration exists
	var toReg models.Registration
	if fromReg.ID != "" {
		toReg, err = dbClient.RegistrationById(fromReg.ID)
	} else if fromReg.Name != "" {
		toReg, err = dbClient.RegistrationByName(fromReg.Name)
	} else {
		http.Error(w, "Need id or name", http.StatusBadRequest)
		return
	}

	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query update registration. Error: %s", err.Error()))
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return
	}

	if fromReg.Name != "" {
		toReg.Name = fromReg.Name
	}
	if fromReg.Addressable.Name != "" {
		toReg.Addressable = fromReg.Addressable
	}
	if fromReg.Format != "" {
		toReg.Format = fromReg.Format
	}
	if fromReg.Filter.DeviceIDs != nil {
		toReg.Filter.DeviceIDs = fromReg.Filter.DeviceIDs
	}
	if fromReg.Filter.ValueDescriptorIDs != nil {
		toReg.Filter.ValueDescriptorIDs = fromReg.Filter.ValueDescriptorIDs
	}
	if fromReg.Encryption.Algo != "" {
		toReg.Encryption = fromReg.Encryption
	}
	if fromReg.Compression != "" {
		toReg.Compression = fromReg.Compression
	}
	if fromReg.Destination != "" {
		toReg.Destination = fromReg.Destination
	}

	// In order to know if 'enable' parameter have been sent or not, we unmarshal again
	// the registration in a map[string] and then check if the parameter is present or not
	var objmap map[string]*json.RawMessage
	json.Unmarshal(data, &objmap)
	if objmap["enable"] != nil {
		toReg.Enable = fromReg.Enable
	}

	if valid, err := toReg.Validate(); !valid {
		LoggingClient.Error(fmt.Sprintf("Failed to validate registrations fields: %X. Error: %s", data, err.Error()))
		http.Error(w, "Could not validate json fields", http.StatusBadRequest)
		return
	}

	err = dbClient.UpdateRegistration(toReg)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query update registration. Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	notifyUpdatedRegistrations(models.NotifyUpdate{Name: toReg.Name,
		Operation: "update"})

	w.Header().Set("Content-Type", applicationJson)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func delRegByID(w http.ResponseWriter, r *http.Request) {

	// URL parameters
	vars := mux.Vars(r)
	id := vars["id"]

	// Read the registration, the registration name is needed to
	// notify distro of the deletion
	reg, err := dbClient.RegistrationById(id)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query by id: %s. Error: %s", id, err.Error()))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = dbClient.DeleteRegistrationById(id)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query by id: %s. Error: %s", id, err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	notifyUpdatedRegistrations(models.NotifyUpdate{Name: reg.Name,
		Operation: "delete"})

	w.Header().Set("Content-Type", applicationJson)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func delRegByName(w http.ResponseWriter, r *http.Request) {
	// URL parameters
	vars := mux.Vars(r)
	name := vars["name"]

	err := dbClient.DeleteRegistrationByName(name)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query by name: %s. Error: %s", name, err.Error()))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	notifyUpdatedRegistrations(models.NotifyUpdate{Name: name,
		Operation: "delete"})

	w.Header().Set("Content-Type", applicationJson)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func notifyUpdatedRegistrations(update models.NotifyUpdate) {
	go func() {
		err := dc.NotifyRegistrations(update, context.Background())
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("error from distro: %s", err.Error()))
		}
	}()
}
