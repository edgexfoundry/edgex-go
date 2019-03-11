/*******************************************************************************
 * Copyright 2018 Dell Technologies Inc.
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
 *******************************************************************************/

package agent

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
)

func LoadRestRoutes() *mux.Router {
	r := mux.NewRouter()
	b := r.PathPrefix("/api/v1").Subrouter()

	b.HandleFunc("/operation", operationHandler).Methods(http.MethodPost)
	b.HandleFunc("/config/{services}", configHandler).Methods(http.MethodGet)
	b.HandleFunc("/metrics/{services}", metricsHandler).Methods(http.MethodGet)

	// Health Resource
	// /api/v1/health
	b.HandleFunc("/health/{services}", healthHandler).Methods(http.MethodGet)

	// Ping Resource
	// /api/v1/ping
	b.HandleFunc("/ping", pingHandler).Methods(http.MethodGet)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)

	return r
}

func operationHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}
	o := models.Operation{}
	err = o.UnmarshalJSON(b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("error during decoding: %v", err.Error())
		return
	} else if o.Action == "" {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	switch o.Action {

	// Call the appropriate internal function (respectively, to stop, start, or restart the service(s)).
	case STOP:
		InvokeOperation(STOP, o.Services)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Done. Stopped the requested services."))
		break

	case START:
		InvokeOperation(START, o.Services)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Done. Started the requested services."))
		break

	case RESTART:
		InvokeOperation(RESTART, o.Services)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Done. Restarted the requested services."))
		break

	default:
		LoggingClient.Warn(o.Action)
	}
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	LoggingClient.Debug("service names: ", vars)

	list := vars["services"]
	var services []string
	services = strings.Split(list, ",")

	ctx := r.Context()
	send, err := getConfig(services, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
		return
	}

	w.Header().Add("Content-Type", "application/json")
	encode(send, w)
	return
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	LoggingClient.Debug("service names: ", vars)

	list := vars["services"]
	var services []string
	services = strings.Split(list, ",")

	ctx := r.Context()
	send, err := getMetrics(services, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
		return
	}

	w.Header().Add("Content-Type", "application/json")
	encode(send, w)
	return
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	LoggingClient.Debug("health status data requested for services: ", vars)

	list := vars["services"]
	var services []string
	services = strings.Split(list, ",")

	send, err := getHealth(services)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("could not retrieve health status: %v: ", err.Error())
		return
	}

	w.Header().Add("Content-Type", "application/json")
	encode(send, w)
	return
}

// Helper function for encoding things for returning from REST calls
func encode(i interface{}, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	err := enc.Encode(i)

	if err != nil {
		LoggingClient.Error("error during encoding: %v", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
