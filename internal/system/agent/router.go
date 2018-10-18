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
	"net/http"
	"github.com/gorilla/mux"
	"fmt"
	"encoding/json"
	"strings"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

func LoadRestRoutes() *mux.Router {
	r := mux.NewRouter()
	b := r.PathPrefix("/api/v1").Subrouter()

	b.HandleFunc("/operation", operationHandler).Methods(http.MethodPost)
	b.HandleFunc("/config/{services}", configHandler).Methods(http.MethodGet)
	b.HandleFunc("/metrics/{services}", metricsHandler).Methods(http.MethodGet)

	// Ping Resource
	// /api/v1/ping
	b.HandleFunc("/ping", pingHandler).Methods(http.MethodGet)

	return r
}

func operationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	// TODO: Work with parsing mux.Vars(r) and assigning to vars.
	var services []string
	vars := mux.Vars(r)
	action := vars["action"]

	var params []string
	var o models.Operation
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&o)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error decoding operation: " + err.Error())
		return
	}

	action = o.Action
	services = o.Services
	params = o.Params

	switch action {

	// Make asynchronous goroutine call(s) to the appropriate internal function (respectively, to stop, start, or restart the service(s).
	case STOP:
		InvokeOperation(STOP, services, params)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Done. Stopped the requested services."))
		break

	case START:
		InvokeOperation(START, services, params)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Done. Started the requested services."))
		break

	case RESTART:
		// First, stop the requested services.
		InvokeOperation(STOP, services, params)
		// Second, start the requested services (thereby effectively restarting those services).
		InvokeOperation(START, services, params)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Done. Restarted the requested services."))
		break

	default:
		LoggingClient.Info(fmt.Sprintf(">> Unknown action %v\n", action))
	}
}

func configHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	LoggingClient.Debug(fmt.Sprintf("It is for these micro-service that their configuration data has been requested: %v", vars))

	list := vars["services"]
	var services []string
	services = strings.Split(list, ",")

	var send = ConfigRespMap{}
	send, _ = getConfig(services)

	w.Header().Add("Content-Type", "application/json")
	encode(send, w)
	return
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	LoggingClient.Debug(fmt.Sprintf("It is for these micro-service that their metrics data has been requested: %v", vars))

	list := vars["services"]
	var services []string
	services = strings.Split(list, ",")

	send, _ := getMetrics(services)

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
		LoggingClient.Error("Error encoding the data: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
