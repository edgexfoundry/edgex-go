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
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/core/command/config"
	container "github.com/edgexfoundry/edgex-go/internal/core/command/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	bootstrapContainer "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
)

func LoadRestRoutes(dic *di.Container) *mux.Router {
	r := mux.NewRouter()

	// Ping Resource
	r.HandleFunc(clients.ApiPingRoute, pingHandler).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(clients.ApiConfigRoute, func(w http.ResponseWriter, r *http.Request) {
		configHandler(w, bootstrapContainer.LoggingClientFrom(dic.Get), container.ConfigurationFrom(dic.Get))
	}).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(clients.ApiMetricsRoute, func(w http.ResponseWriter, r *http.Request) {
		metricsHandler(w, bootstrapContainer.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodGet)

	// Version
	r.HandleFunc(clients.ApiVersionRoute, pkg.VersionHandler).Methods(http.MethodGet)

	b := r.PathPrefix(clients.ApiBase).Subrouter()

	loadDeviceRoutes(b, dic)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)

	return r
}

func loadDeviceRoutes(b *mux.Router, dic *di.Container) {

	b.HandleFunc("/device", func(w http.ResponseWriter, r *http.Request) {
		restGetAllCommands(
			w,
			r,
			bootstrapContainer.DBClientFrom(dic.Get),
			container.MetadataDeviceClientFrom(dic.Get),
			container.ConfigurationFrom(dic.Get))
	}).Methods(http.MethodGet)

	d := b.PathPrefix("/" + DEVICE).Subrouter()

	// /api/<version>/device
	d.HandleFunc("/{"+ID+"}", func(w http.ResponseWriter, r *http.Request) {
		restGetCommandsByDeviceID(
			w,
			r,
			bootstrapContainer.DBClientFrom(dic.Get),
			container.MetadataDeviceClientFrom(dic.Get),
			container.ConfigurationFrom(dic.Get))
	}).Methods(http.MethodGet)
	d.HandleFunc("/{"+ID+"}/"+COMMAND+"/{"+COMMANDID+"}", func(w http.ResponseWriter, r *http.Request) {
		restGetDeviceCommandByCommandID(
			w,
			r,
			bootstrapContainer.LoggingClientFrom(dic.Get),
			bootstrapContainer.DBClientFrom(dic.Get),
			container.MetadataDeviceClientFrom(dic.Get))
	}).Methods(http.MethodGet)
	d.HandleFunc("/{"+ID+"}/"+COMMAND+"/{"+COMMANDID+"}", func(w http.ResponseWriter, r *http.Request) {
		restPutDeviceCommandByCommandID(
			w,
			r,
			bootstrapContainer.LoggingClientFrom(dic.Get),
			bootstrapContainer.DBClientFrom(dic.Get),
			container.MetadataDeviceClientFrom(dic.Get))
	}).Methods(http.MethodPut)

	// /api/<version>/device/name
	dn := d.PathPrefix("/" + NAME).Subrouter()

	dn.HandleFunc("/{"+NAME+"}", func(w http.ResponseWriter, r *http.Request) {
		restGetCommandsByDeviceName(
			w,
			r,
			bootstrapContainer.DBClientFrom(dic.Get),
			container.MetadataDeviceClientFrom(dic.Get),
			container.ConfigurationFrom(dic.Get))
	}).Methods(http.MethodGet)
	dn.HandleFunc("/{"+NAME+"}/"+COMMAND+"/{"+COMMANDNAME+"}", func(w http.ResponseWriter, r *http.Request) {
		restGetDeviceCommandByNames(
			w,
			r,
			bootstrapContainer.LoggingClientFrom(dic.Get),
			bootstrapContainer.DBClientFrom(dic.Get),
			container.MetadataDeviceClientFrom(dic.Get))
	}).Methods(http.MethodGet)
	dn.HandleFunc("/{"+NAME+"}/"+COMMAND+"/{"+COMMANDNAME+"}", func(w http.ResponseWriter, r *http.Request) {
		restPutDeviceCommandByNames(
			w,
			r,
			bootstrapContainer.LoggingClientFrom(dic.Get),
			bootstrapContainer.DBClientFrom(dic.Get),
			container.MetadataDeviceClientFrom(dic.Get))
	}).Methods(http.MethodPut)
}

// Test if the service is working
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(clients.ContentType, clients.ContentTypeText)
	w.Write([]byte("pong"))
}

func configHandler(w http.ResponseWriter, loggingClient logger.LoggingClient, configuration *config.ConfigurationStruct) {
	pkg.Encode(configuration, w, loggingClient)
}

func metricsHandler(w http.ResponseWriter, loggingClient logger.LoggingClient) {
	s := telemetry.NewSystemUsage()

	pkg.Encode(s, w, loggingClient)

	return
}
