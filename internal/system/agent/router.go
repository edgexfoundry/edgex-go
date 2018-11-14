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
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
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
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(fmt.Sprintf("unable to read request body (%s)", err.Error()))
		return
	}
	o := models.Operation{}
	err = o.UnmarshalJSON(b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error decoding operation: " + err.Error())
		return
	} else if o.Action == "" {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("action is required")
		return
	}

	// invoke the operation and check the result
	err = InvokeOperation(o.Action, o.Services, o.Parameters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error invoking operation: " + err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	var msg string

	switch o.Action {
	case STOP:
		msg = "Done. Stopped the requested services."
	case START:
		msg = "Done. Started the requested services."
	case RESTART:
		msg = "Done. Restarted the requested services."
	case ENABLE:
		msg = "Done. Enabled the requested services."
	case DISABLE:
		msg = "Done. Disabled the requested services."
	}
	// note that an unsupported action is handled with an error returned from InvokeOperation
	w.Write([]byte(msg))
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
