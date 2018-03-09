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
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/edgexfoundry/edgex-go/export"
	"github.com/edgexfoundry/edgex-go/export/mongo"
	"github.com/go-zoo/bone"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

const (
	distroPort int = 48070
)

func getRegByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	id := bone.GetValue(r, "id")

	s := repo.Session.Copy()
	defer s.Close()
	c := s.DB(mongo.DBName).C(mongo.CollectionName)

	reg := export.Registration{}
	if err := c.Find(bson.M{"id": id}).One(&reg); err != nil {
		logger.Error("Failed to query by id", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}

	res, err := json.Marshal(reg)
	if err != nil {
		logger.Error("Failed to query by id", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(res))
}

func getRegList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	t := bone.GetValue(r, "type")

	var list []string

	switch t {
	case "algorithms":
		list = append(list, export.EncNone)
		list = append(list, export.EncAes)
	case "compressions":
		list = append(list, export.CompNone)
		list = append(list, export.CompGzip)
		list = append(list, export.CompZip)
	case "formats":
		list = append(list, export.FormatJSON)
		list = append(list, export.FormatXML)
	case "destinations":
		list = append(list, export.DestMQTT)
		list = append(list, export.DestRest)
	default:
		logger.Error("Unknown type: " + t)
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Unknown type: "+t)
		return
	}

	res, err := json.Marshal(list)
	if err != nil {
		logger.Error("Failed to generate json", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(res))
}

func getAllReg(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	s := repo.Session.Copy()
	defer s.Close()
	c := s.DB(mongo.DBName).C(mongo.CollectionName)

	reg := []export.Registration{}
	if err := c.Find(nil).All(&reg); err != nil {
		logger.Error("Failed to query all registrations", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}

	res, err := json.Marshal(reg)
	if err != nil {
		logger.Error("Failed to query all registrations", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(res))
}

func getRegByName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	name := bone.GetValue(r, "name")

	s := repo.Session.Copy()
	defer s.Close()
	c := s.DB(mongo.DBName).C(mongo.CollectionName)

	reg := export.Registration{}
	if err := c.Find(bson.M{"name": name}).One(&reg); err != nil {
		logger.Error("Failed to query by name", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}

	res, err := json.Marshal(reg)
	if err != nil {
		logger.Error("Failed to query by name", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(res))
}

func addReg(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error("Failed to query add registration", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	reg := export.Registration{}
	if err := json.Unmarshal(data, &reg); err != nil {
		logger.Error("Failed to query add registration", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	if !reg.Validate() {
		logger.Error("Failed to validate registrations fields", zap.ByteString("data", data))
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Could not validate json fields")
		return
	}

	s := repo.Session.Copy()
	defer s.Close()
	c := s.DB(mongo.DBName).C(mongo.CollectionName)

	count, err := c.Find(bson.M{"name": reg.Name}).Count()
	if err != nil {
		logger.Error("Failed to query add registration", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}
	if count != 0 {
		logger.Error("Username already taken: " + reg.Name)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := c.Insert(reg); err != nil {
		logger.Error("Failed to query add registration", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	notifyUpdatedRegistrations(export.NotifyUpdate{Name: reg.Name,
		Operation: "add"})
}

func updateReg(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error("Failed to query update registration", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		logger.Error("Failed to query update registration", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
	}

	s := repo.Session.Copy()
	defer s.Close()
	c := s.DB(mongo.DBName).C(mongo.CollectionName)

	name := body["name"]
	query := bson.M{"name": name}
	update := bson.M{"$set": body}

	if err := c.Update(query, update); err != nil {
		logger.Error("Failed to query update registration", zap.Error(err))
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	notifyUpdatedRegistrations(export.NotifyUpdate{Name: name.(string),
		Operation: "update"})
}

func delRegByID(w http.ResponseWriter, r *http.Request) {
	id := bone.GetValue(r, "id")

	s := repo.Session.Copy()
	defer s.Close()
	c := s.DB(mongo.DBName).C(mongo.CollectionName)

	// Read the registration from mongo, the registration name is needed to
	// notify distro of the deletion
	reg := export.Registration{}
	if err := c.Find(bson.M{"id": id}).One(&reg); err != nil {
		logger.Error("Failed to query by id", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}

	if err := c.Remove(bson.M{"id": id}); err != nil {
		logger.Error("Failed to query by id", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	notifyUpdatedRegistrations(export.NotifyUpdate{Name: reg.Name,
		Operation: "delete"})
}

func delRegByName(w http.ResponseWriter, r *http.Request) {
	name := bone.GetValue(r, "name")

	s := repo.Session.Copy()
	defer s.Close()
	c := s.DB(mongo.DBName).C(mongo.CollectionName)

	if err := c.Remove(bson.M{"name": name}); err != nil {
		logger.Error("Failed to query by name", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	notifyUpdatedRegistrations(export.NotifyUpdate{Name: name,
		Operation: "delete"})
}

func notifyUpdatedRegistrations(update export.NotifyUpdate) {
	go func() {
		client := &http.Client{}
		url := "http://" + cfg.DistroHost + ":" + strconv.Itoa(distroPort) +
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
