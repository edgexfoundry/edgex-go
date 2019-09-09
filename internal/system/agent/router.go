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

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/executor"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/gorilla/mux"
)

func LoadRestRoutes(instance *Instance, metricsImpl interfaces.Metrics) *mux.Router {
	r := mux.NewRouter()

	b := r.PathPrefix("/api/v1").Subrouter()
	b.HandleFunc("/operation", func(w http.ResponseWriter, r *http.Request) { operationHandler(w, r, instance) }).Methods(http.MethodPost)
	b.HandleFunc("/config/{services}", func(w http.ResponseWriter, r *http.Request) { configHandler(w, r, instance) }).Methods(http.MethodGet)
	b.HandleFunc("/metrics/{services}", func(w http.ResponseWriter, r *http.Request) { metricsHandler(w, r, instance, metricsImpl) }).Methods(http.MethodGet)
	b.HandleFunc("/health/{services}", func(w http.ResponseWriter, r *http.Request) { healthHandler(w, r, instance) }).Methods(http.MethodGet)
	b.HandleFunc("/ping", pingHandler).Methods(http.MethodGet)

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
func metricsHandler(w http.ResponseWriter, r *http.Request, instance *Instance, metricsImpl interfaces.Metrics) {
	instance.LoggingClient.Debug("retrieved service names")

	vars := mux.Vars(r)
	pkg.Encode(metricsImpl.Get(strings.Split(vars["services"], ","), r.Context()), w, instance.LoggingClient)
}

// operationHandler implements a controller to execute a start/stop/restart operation request.
func operationHandler(w http.ResponseWriter, r *http.Request, instance *Instance) {
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		instance.LoggingClient.Error(err.Error())
		return
	}

	o := models.Operation{}
	if err = o.UnmarshalJSON(b); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		instance.LoggingClient.Error("error during decoding: %s", err.Error())
		return
	}

	operation := executor.NewOperations(
		executor.CommandExecutor,
		instance.LoggingClient,
		instance.Configuration.ExecutorPath)
	pkg.Encode(operation.Do(o.Services, o.Action), w, instance.LoggingClient)
}

func configHandler(w http.ResponseWriter, r *http.Request, instance *Instance) {
	vars := mux.Vars(r)
	instance.LoggingClient.Debug("retrieved service names")

	list := vars["services"]
	var services []string
	services = strings.Split(list, ",")

	ctx := r.Context()
	send, err := getConfig(
		services,
		ctx,
		instance.LoggingClient,
		instance.GenClients,
		instance.Configuration.Clients,
		instance.RegistryClient,
		instance.Configuration.Service.Protocol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		instance.LoggingClient.Error(err.Error())
		return
	}
	pkg.Encode(send, w, instance.LoggingClient)
}

func healthHandler(w http.ResponseWriter, r *http.Request, instance *Instance) {
	vars := mux.Vars(r)
	instance.LoggingClient.Debug("health status data requested")

	list := vars["services"]
	var services []string
	services = strings.Split(list, ",")

	send, err := getHealth(services, instance.LoggingClient, instance.RegistryClient)
	if err != nil {
		instance.LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pkg.Encode(send, w, instance.LoggingClient)
}
