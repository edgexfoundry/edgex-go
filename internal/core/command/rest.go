/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
 *******************************************************************************/
package command

import (
	"encoding/json"
	"net/http"
	"runtime"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/gorilla/mux"
)

func LoadRestRoutes() http.Handler {
	r := mux.NewRouter()

	// Ping Resource
	r.HandleFunc(clients.ApiPingRoute, pingHandler).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(clients.ApiConfigRoute, configHandler).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(clients.ApiMetricsRoute, metricsHandler).Methods(http.MethodGet)

	b := r.PathPrefix(clients.ApiBase).Subrouter()

	loadDeviceRoutes(b)
	return r
}

func loadDeviceRoutes(b *mux.Router) {

	b.HandleFunc("/device", restGetAllCommands).Methods(http.MethodGet)

	d := b.PathPrefix("/" + DEVICE).Subrouter()

	// /api/<version>/device
	d.HandleFunc("/{"+ID+"}", restGetCommandsByDeviceID).Methods(http.MethodGet)
	d.HandleFunc("/{"+ID+"}/"+COMMAND+"/{"+COMMANDID+"}", restGetDeviceCommandByCommandID).Methods(http.MethodGet)
	d.HandleFunc("/{"+ID+"}/"+COMMAND+"/{"+COMMANDID+"}", restPutDeviceCommandByCommandID).Methods(http.MethodPut)
	d.HandleFunc("/{"+ID+"}/"+URLADMINSTATE+"/{"+ADMINSTATE+"}", restPutDeviceAdminState).Methods(http.MethodPut)
	d.HandleFunc("/{"+ID+"}/"+OPSTATE+"/{"+OPSTATE+"}", restPutDeviceOpState).Methods(http.MethodPut)

	// /api/<version>/device/name
	dn := d.PathPrefix("/" + NAME).Subrouter()

	dn.HandleFunc("/{"+NAME+"}", restGetCommandsByDeviceName).Methods(http.MethodGet)
	dn.HandleFunc("/{"+NAME+"}/"+URLADMINSTATE+"/{"+ADMINSTATE+"}", restPutDeviceAdminStateByDeviceName).Methods(http.MethodPut)
	dn.HandleFunc("/{"+NAME+"}/"+OPSTATE+"/{"+OPSTATE+"}", restPutDeviceOpStateByDeviceName).Methods(http.MethodPut)
}

// Respond with PINGRESPONSE to see if the service is alive
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(CONTENTTYPE, TEXTPLAIN)
	w.Write([]byte(PINGRESPONSE))
}

func configHandler(w http.ResponseWriter, _ *http.Request) {
	encode(Configuration, w)
}

func metricsHandler(w http.ResponseWriter, _ *http.Request) {
	var t internal.Telemetry

	// The micro-service is to be considered the System Of Record (SOR) in terms of accurate information.
	// Fetch metrics for the command service.
	var rtm runtime.MemStats

	// Read full memory stats
	runtime.ReadMemStats(&rtm)

	// Miscellaneous memory stats
	t.Alloc = rtm.Alloc
	t.TotalAlloc = rtm.TotalAlloc
	t.Sys = rtm.Sys
	t.Mallocs = rtm.Mallocs
	t.Frees = rtm.Frees

	// Live objects = Mallocs - Frees
	t.LiveObjects = t.Mallocs - t.Frees

	encode(t, w)

	return
}

// Helper function for encoding things for returning from REST calls
func encode(i interface{}, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	err := enc.Encode(i)
	// Problems encoding
	if err != nil {
		LoggingClient.Error("Error encoding the data: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
