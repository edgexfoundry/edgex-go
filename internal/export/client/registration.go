//
// Copyright (c) 2017
// Mainflux
// Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/edgexfoundry/edgex-go/internal/export"
	"github.com/go-zoo/bone"
	"go.uber.org/zap"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
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
	id := bone.GetValue(r, "id")

	reg, err := dbc.RegistrationById(id)
	if err != nil {
		logger.Error("Failed to query by id", zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", applicationJson)
	json.NewEncoder(w).Encode(&reg)
}

func getRegList(w http.ResponseWriter, r *http.Request) {
	t := bone.GetValue(r, "type")

	var list []string

	switch t {
	case typeAlgorithms:
		list = append(list, export.EncNone)
		list = append(list, export.EncAes)
	case typeCompressions:
		list = append(list, export.CompNone)
		list = append(list, export.CompGzip)
		list = append(list, export.CompZip)
	case typeFormats:
		list = append(list, export.FormatJSON)
		list = append(list, export.FormatXML)
	case typeDestinations:
		list = append(list, export.DestMQTT)
		list = append(list, export.DestRest)
	default:
		logger.Error("Unknown type: " + t)
		http.Error(w, "Unknown type: "+t, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", applicationJson)
	json.NewEncoder(w).Encode(&list)
}

func getAllReg(w http.ResponseWriter, r *http.Request) {
	reg, err := dbc.Registrations()
	if err != nil {
		logger.Error("Failed to query all registrations", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", applicationJson)
	json.NewEncoder(w).Encode(&reg)
}

func getRegByName(w http.ResponseWriter, r *http.Request) {
	name := bone.GetValue(r, "name")

	reg, err := dbc.RegistrationByName(name)
	if err != nil {
		logger.Error("Failed to query by name", zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", applicationJson)
	json.NewEncoder(w).Encode(&reg)
}

func addReg(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error("Failed to query add registration", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reg := export.Registration{}
	if err := json.Unmarshal(data, &reg); err != nil {
		logger.Error("Failed to query add registration", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if valid, err := reg.Validate(); !valid {
		logger.Error("Failed to validate registrations fields", zap.ByteString("data", data), zap.Error(err))
		http.Error(w, "Could not validate json fields", http.StatusBadRequest)
		return
	}

	_, err = dbc.RegistrationByName(reg.Name)
	if err == nil {
		logger.Error("Name already taken: " + reg.Name)
		http.Error(w, "Name already taken", http.StatusBadRequest)
		return
	} else if err != db.ErrNotFound {
		logger.Error("Failed to query add registration", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = dbc.AddRegistration(&reg)
	if err != nil {
		logger.Error("Failed to query add registration", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	notifyUpdatedRegistrations(export.NotifyUpdate{Name: reg.Name,
		Operation: "add"})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(reg.ID.Hex()))
}

func updateReg(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error("Failed to read update registration", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var fromReg export.Registration
	if err := json.Unmarshal(data, &fromReg); err != nil {
		logger.Error("Failed to unmarshal update registration", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the registration exists
	var toReg export.Registration
	if fromReg.ID != "" {
		toReg, err = dbc.RegistrationById(fromReg.ID.Hex())
	} else if fromReg.Name != "" {
		toReg, err = dbc.RegistrationByName(fromReg.Name)
	} else {
		http.Error(w, "Need id or name", http.StatusBadRequest)
		return
	}

	if err != nil {
		logger.Error("Failed to query update registration", zap.Error(err))
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
		logger.Error("Failed to validate registrations fields", zap.ByteString("data", data), zap.Error(err))
		http.Error(w, "Could not validate json fields", http.StatusBadRequest)
		return
	}

	err = dbc.UpdateRegistration(toReg)
	if err != nil {
		logger.Error("Failed to query update registration", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	notifyUpdatedRegistrations(export.NotifyUpdate{Name: toReg.Name,
		Operation: "update"})

	w.Header().Set("Content-Type", applicationJson)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func delRegByID(w http.ResponseWriter, r *http.Request) {
	id := bone.GetValue(r, "id")

	// Read the registration, the registration name is needed to
	// notify distro of the deletion
	reg, err := dbc.RegistrationById(id)
	if err != nil {
		logger.Error("Failed to query by id", zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = dbc.DeleteRegistrationById(id)
	if err != nil {
		logger.Error("Failed to query by id", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	notifyUpdatedRegistrations(export.NotifyUpdate{Name: reg.Name,
		Operation: "delete"})

	w.Header().Set("Content-Type", applicationJson)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func delRegByName(w http.ResponseWriter, r *http.Request) {
	name := bone.GetValue(r, "name")

	err := dbc.DeleteRegistrationByName(name)
	if err != nil {
		logger.Error("Failed to query by name", zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	notifyUpdatedRegistrations(export.NotifyUpdate{Name: name,
		Operation: "delete"})

	w.Header().Set("Content-Type", applicationJson)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func notifyUpdatedRegistrations(update export.NotifyUpdate) {
	go func() {
		client := &http.Client{}
		url := "http://" + configuration.DistroHost + ":" + strconv.Itoa(configuration.DistroPort) +
			"/api/v1/notify/registrations"

		data, err := json.Marshal(update)
		if err != nil {
			logger.Error("Error generating update json", zap.Error(err))
			return
		}

		req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer([]byte(data)))
		if err != nil {
			logger.Error("Error creating http request")
			return
		}
		_, err = client.Do(req)
		if err != nil {
			logger.Error("Error notifying updated registrations to distro", zap.String("url", url))
		}
	}()
}
