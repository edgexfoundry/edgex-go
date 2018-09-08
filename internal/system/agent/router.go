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
)

func LoadRestRoutes() *mux.Router {
	r := mux.NewRouter()
	b := r.PathPrefix("/api/v1").Subrouter()

	// Notifications
	b.HandleFunc("/operation", operationServiceHandler).Methods(http.MethodPost)
	b.HandleFunc("/config", configHandler).Methods(http.MethodGet)
	b.HandleFunc("/metric", metricHandler).Methods(http.MethodGet)

	// Ping Resource
	// /api/v1/ping
	b.HandleFunc("/ping", pingHandler).Methods(http.MethodGet)

	return r
}

func operationServiceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	// TODO: Work with parsing mux.Vars(r) and assigning to vars.
	//vars := mux.Vars(r)
	//action := vars["ops"]
	//params :vars["params"]
	//services := vars["services"]
	var action string
	var params map[string]string
	var services []string

	switch action {

	// Make asynchronous call(s) to the appropriate internal function (to stop, start, or restart the service(s).

	case START:
		go invokeAction(START, services, params)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Response"))
		break

	case STOP:
		go invokeAction(STOP, services, params)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Response"))
		break

	case RESTART:
		// First, stop the requested services.
		go invokeAction(STOP, services, params)
		// Second, start the requested services (thereby effectively restarting those services).
		go invokeAction(START, services, params)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Response"))
		break

	default:
		LoggingClient.Info(fmt.Sprintf(">> Unknown action %v\n", action))
	}
}

func configHandler(w http.ResponseWriter, r *http.Request) {

	// Example Request: { “services”: [“edgex-core-data”, “edgex-core-metadata”, …] }
	if r.Body != nil {
		defer r.Body.Close()
	}

	// TODO: Work with parsing mux.Vars(r) and assigning to vars.
	//vars := mux.Vars(r)
	//services := vars["services"]
	var services []string

	// Make asynchronous call to the microservices' API for configuration.
	go getConfig(services)

	// Example Response:
	/*
          [
          {
             "service":"edgex-core-data",
             "config":[
                "port":48080,
                "loggingLevel":"debug"         …
             ]
          },
          {
             "service":"edgex-core-metdata",
             "config":[
                "port":48081,
                "loggingLevel":"error"         …
             ]
          }
          ]
	*/
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Example Response..."))

}

func metricHandler(w http.ResponseWriter, r *http.Request) {

	// Example Request: {“metrics”:[“memory”, “CPU”], “services”: [“edgex-core-data”, “edgex-core-metadata”, …] }
	if r.Body != nil {
		defer r.Body.Close()
	}

	// TODO: Work with parsing mux.Vars(r) and assigning to vars.
	//vars := mux.Vars(r)
	//metrics := vars["metrics"]
	//services := vars["services"]
	var services []string
	var metrics []string

	// Make asynchronous call to the microservices' API for metrics.
	go getMetric(services, metrics)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Response"))

	/* Example Response:
	    [
            {
               "service":"edgex-core-data",
               "metrics":[
                  "memory":"34MB",
                  "CPU":"3%"
               ]
            },
            {
               "service":"edgex-core-metdata",
               "metrics":[
                  "memory":"31MB",
                  "CPU":"2%"
               ]
            },
            …
        ]
	*/
}
