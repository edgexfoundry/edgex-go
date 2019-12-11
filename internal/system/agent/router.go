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
	"github.com/edgexfoundry/edgex-go/internal/system/agent/container"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	requests "github.com/edgexfoundry/go-mod-core-contracts/requests/configuration"

	"github.com/edgexfoundry/go-mod-registry/registry"

	"github.com/gorilla/mux"
)

func LoadRestRoutes(dic *di.Container) *mux.Router {
	r := mux.NewRouter()

	b := r.PathPrefix("/api/v1").Subrouter()

	b.HandleFunc(
		"/operation",
		func(w http.ResponseWriter, r *http.Request) {
			operationHandler(w, r, bootstrapContainer.LoggingClientFrom(dic.Get), container.OperationsFrom(dic.Get))
		}).Methods(http.MethodPost)

	b.HandleFunc(
		"/config/{services}",
		func(w http.ResponseWriter, r *http.Request) {
			getConfigHandler(w, r, bootstrapContainer.LoggingClientFrom(dic.Get), container.GetConfigFrom(dic.Get))
		}).Methods(http.MethodGet)

	b.HandleFunc(
		"/config/{services}",
		func(w http.ResponseWriter, r *http.Request) {
			setConfigHandler(w, r, bootstrapContainer.LoggingClientFrom(dic.Get), container.SetConfigFrom(dic.Get))
		}).Methods(http.MethodPut)

	b.HandleFunc(
		"/metrics/{services}",
		func(w http.ResponseWriter, r *http.Request) {
			metricsHandler(w, r, bootstrapContainer.LoggingClientFrom(dic.Get), container.MetricsFrom(dic.Get))
		}).Methods(http.MethodGet)

	b.HandleFunc(
		"/health/{services}",
		func(w http.ResponseWriter, r *http.Request) {
			healthHandler(w, r, bootstrapContainer.LoggingClientFrom(dic.Get), bootstrapContainer.RegistryFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/ping",
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set(clients.ContentType, clients.ContentTypeText)
			_, _ = w.Write([]byte("pong"))
		}).Methods(http.MethodGet)

	r.HandleFunc(clients.ApiVersionRoute, pkg.VersionHandler).Methods(http.MethodGet)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)

	return r
}

// metricsHandler implements a controller to execute a metrics request.
func metricsHandler(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	metricsImpl interfaces.Metrics) {

	loggingClient.Debug("retrieved service names")

	vars := mux.Vars(r)
	pkg.Encode(metricsImpl.Get(r.Context(), strings.Split(vars["services"], ",")), w, loggingClient)
}

// operationHandler implements a controller to execute a start/stop/restart operation request.
func operationHandler(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	operationsImpl interfaces.Operations) {

	defer func() { _ = r.Body.Close() }()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		loggingClient.Error(err.Error())
		return
	}

	o := models.Operation{}
	if err = o.UnmarshalJSON(b); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		loggingClient.Error("error during decoding: %s", err.Error())
		return
	}

	if len(o.Services) == 0 || len(o.Action) == 0 {
		const errorMessage = "incorrect or malformed body was passed in with the request"
		http.Error(w, errorMessage, http.StatusBadRequest)
		loggingClient.Error(errorMessage)
		return
	}

	pkg.Encode(operationsImpl.Do(o.Services, o.Action), w, loggingClient)
}

// getConfigHandler implements a controller to execute a get configuration request.
func getConfigHandler(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,

	getConfigImpl interfaces.GetConfig) {
	vars := mux.Vars(r)
	loggingClient.Debug("retrieved service names")

	pkg.Encode(getConfigImpl.Do(r.Context(), strings.Split(vars["services"], ",")), w, loggingClient)
}

// setConfigHandler implements a controller to execute a set configuration request.
func setConfigHandler(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	setConfigImpl interfaces.SetConfig) {

	defer func() { _ = r.Body.Close() }()

	vars := mux.Vars(r)
	loggingClient.Debug("retrieved service names")

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		loggingClient.Error(err.Error())
		return
	}

	sc := requests.SetConfigRequest{}
	if err = sc.UnmarshalJSON(b); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		loggingClient.Error("error during decoding")
		return
	}

	pkg.Encode(setConfigImpl.Do(strings.Split(vars["services"], ","), sc), w, loggingClient)
}

// healthHandler implements a controller to execute a get health status request.
func healthHandler(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	registryClient registry.Client) {

	vars := mux.Vars(r)
	loggingClient.Debug("health status data requested")

	list := vars["services"]
	var services []string
	services = strings.Split(list, ",")

	send, err := getHealth(services, registryClient)
	if err != nil {
		loggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pkg.Encode(send, w, loggingClient)
}
