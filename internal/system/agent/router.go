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
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
)

func LoadRestRoutes() *mux.Router {
	r := mux.NewRouter()
	b := r.PathPrefix("/api/v1").Subrouter()

	b.HandleFunc("/"+operation, operationHandler).Methods(http.MethodPost)
	b.HandleFunc("/"+configuration+"/{"+services+"}", configHandler).Methods(http.MethodGet)
	b.HandleFunc("/"+metrics+"/{"+services+"}", metricsHandler).Methods(http.MethodGet)

	// Health Resource
	// /api/v1/health
	b.HandleFunc("/"+health+"/{"+services+"}", healthHandler).Methods(http.MethodGet)

	// Ping Resource
	// /api/v1/ping
	b.HandleFunc("/"+ping, pingHandler).Methods(http.MethodGet)

	// Version
	r.HandleFunc(clients.ApiVersionRoute, pkg.VersionHandler).Methods(http.MethodGet)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)

	return r
}

// pingHandler implements a controller to execute a ping request.
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong"))
}

// metricsHandler implements a controller to execute a metrics request.
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	LoggingClient.Debug("retrieved service names")

	list := vars["services"]
	var services []string
	services = strings.Split(list, ",")

	send, err := invokeMetrics(services, r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
		return
	}
	pkg.Encode(send, w, LoggingClient)
}

// operationHandler implements a controller to execute a start/stop/restart operation request.
func operationHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	o := models.Operation{}
	if err = o.UnmarshalJSON(b); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("error during decoding: %s", err.Error())
		return
	}

	send, err := invokeOperation(o.Action, o.Services)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
		return
	}

	pkg.Encode(send, w, LoggingClient)
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	LoggingClient.Debug("retrieved service names")

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
	pkg.Encode(send, w, LoggingClient)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	LoggingClient.Debug("health status data requested")

	list := vars["services"]
	var services []string
	services = strings.Split(list, ",")

	send, err := getHealth(services)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pkg.Encode(send, w, LoggingClient)
}
