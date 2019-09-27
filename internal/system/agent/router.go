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
	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	requests "github.com/edgexfoundry/go-mod-core-contracts/requests/configuration"

	"github.com/gorilla/mux"
)

func LoadRestRoutes(
	metricsImpl interfaces.Metrics,
	operationsImpl interfaces.Operations,
	configGetImpl interfaces.GetConfig,
	configSetImpl interfaces.SetConfig) *mux.Router {
	r := mux.NewRouter()

	b := r.PathPrefix("/api/v1").Subrouter()
	b.HandleFunc("/operation", func(w http.ResponseWriter, r *http.Request) { operationHandler(w, r, operationsImpl) }).Methods(http.MethodPost)
	b.HandleFunc("/config/{services}", func(w http.ResponseWriter, r *http.Request) { getConfigHandler(w, r, configGetImpl) }).Methods(http.MethodGet)
	b.HandleFunc("/config/{services}", func(w http.ResponseWriter, r *http.Request) { setConfigHandler(w, r, configSetImpl) }).Methods(http.MethodPut)
	b.HandleFunc("/metrics/{services}", func(w http.ResponseWriter, r *http.Request) { metricsHandler(w, r, metricsImpl) }).Methods(http.MethodGet)
	b.HandleFunc("/health/{services}", healthHandler).Methods(http.MethodGet)
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
func metricsHandler(w http.ResponseWriter, r *http.Request, metricsImpl interfaces.Metrics) {
	LoggingClient.Debug("retrieved service names")

	vars := mux.Vars(r)
	pkg.Encode(metricsImpl.Get(strings.Split(vars["services"], ","), r.Context()), w, LoggingClient)
}

// operationHandler implements a controller to execute a start/stop/restart operation request.
func operationHandler(w http.ResponseWriter, r *http.Request, operationsImpl interfaces.Operations) {
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

	pkg.Encode(operationsImpl.Do(o.Services, o.Action), w, LoggingClient)
}

func getConfigHandler(w http.ResponseWriter, r *http.Request, getConfigImpl interfaces.GetConfig) {
	vars := mux.Vars(r)
	LoggingClient.Debug("retrieved service names")

	pkg.Encode(getConfigImpl.Do(strings.Split(vars["services"], ","), r.Context()), w, LoggingClient)
}

func setConfigHandler(w http.ResponseWriter, r *http.Request, setConfigImpl interfaces.SetConfig) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	LoggingClient.Debug("retrieved service names")

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	sc := requests.SetConfigRequest{}
	if err = sc.UnmarshalJSON(b); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("error during decoding")
		return
	}

	pkg.Encode(setConfigImpl.Do(strings.Split(vars["services"], ","), sc), w, LoggingClient)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	LoggingClient.Debug("health status data requested")

	list := vars["services"]
	var services []string
	services = strings.Split(list, ",")

	send, err := getHealth(services, RegistryClient)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pkg.Encode(send, w, LoggingClient)
}
