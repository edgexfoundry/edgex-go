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

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
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
		LoggingClient.Error("error during decoding")
		return
	} else if len(o.Services) == 0 || len(o.Action) == 0 {
		http.Error(w, "incorrect or malformed body was passed in with the request", http.StatusBadRequest)
		LoggingClient.Error("incorrect or malformed body was passed in with the request")
		return
	}

	switch o.Action {
	case STOP:
		err := InvokeOperation(STOP, o.Services)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Done. Stopped the requested services."))
		}

		break

	case START:
		err := InvokeOperation(START, o.Services)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Done. Started the requested services."))
		}
		break

	case RESTART:
		err := InvokeOperation(RESTART, o.Services)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Done. Restarted the requested services."))
		}
		break

	default:
		LoggingClient.Warn(o.Action)
	}
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

	w.Header().Add(clients.ContentType, clients.ContentTypeJSON)
	encode(send, w)
	return
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	LoggingClient.Debug("retrieved service names")

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

	w.Header().Add(clients.ContentType, clients.ContentTypeJSON)
	encode(send, w)
	return
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

	w.Header().Add(clients.ContentType, clients.ContentTypeJSON)
	encode(send, w)
	return
}

// Helper function for encoding things for returning from REST calls
func encode(i interface{}, w http.ResponseWriter) {
	w.Header().Add(clients.ContentType, clients.ContentTypeJSON)

	enc := json.NewEncoder(w)
	err := enc.Encode(i)

	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
