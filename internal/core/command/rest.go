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
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
)

func LoadRestRoutes() http.Handler {
	r := mux.NewRouter()

	// Ping Resource
	r.HandleFunc(clients.ApiPingRoute, pingHandler).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(clients.ApiConfigRoute, configHandler).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(clients.ApiMetricsRoute, metricsHandler).Methods(http.MethodGet)

	// Version
	r.HandleFunc(clients.ApiVersionRoute, pkg.VersionHandler).Methods(http.MethodGet)

	b := r.PathPrefix(clients.ApiBase).Subrouter()

	loadDeviceRoutes(b)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)

	return r
}

func loadDeviceRoutes(b *mux.Router) {

	b.HandleFunc("/device", restGetAllCommands).Methods(http.MethodGet)

	d := b.PathPrefix("/" + DEVICE).Subrouter()

	// /api/<version>/device
	d.HandleFunc("/{"+ID+"}", restGetCommandsByDeviceID).Methods(http.MethodGet)
	d.HandleFunc("/{"+ID+"}/"+COMMAND+"/{"+COMMANDID+"}", restGetDeviceCommandByCommandID).Methods(http.MethodGet)
	d.HandleFunc("/{"+ID+"}/"+COMMAND+"/{"+COMMANDID+"}", restPutDeviceCommandByCommandID).Methods(http.MethodPut)

	// /api/<version>/device/name
	dn := d.PathPrefix("/" + NAME).Subrouter()

	dn.HandleFunc("/{"+NAME+"}", restGetCommandsByDeviceName).Methods(http.MethodGet)
	dn.HandleFunc("/{"+NAME+"}/"+COMMAND+"/{"+COMMANDNAME+"}", restGetDeviceCommandByNames).Methods(http.MethodGet)
	dn.HandleFunc("/{"+NAME+"}/"+COMMAND+"/{"+COMMANDNAME+"}", restPutDeviceCommandByNames).Methods(http.MethodPut)
}

// Respond with PINGRESPONSE to see if the service is alive
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(CONTENTTYPE, TEXTPLAIN)
	w.Write([]byte(PINGRESPONSE))
}

func configHandler(w http.ResponseWriter, _ *http.Request) {
	pkg.Encode(Configuration, w, LoggingClient)
}

func metricsHandler(w http.ResponseWriter, _ *http.Request) {
	s := telemetry.NewSystemUsage()

	pkg.Encode(s, w, LoggingClient)

	return
}
