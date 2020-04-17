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

	commandContainer "github.com/edgexfoundry/edgex-go/internal/core/command/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	errorContainer "github.com/edgexfoundry/edgex-go/internal/pkg/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"

	"github.com/gorilla/mux"
)

func loadRestRoutes(r *mux.Router, dic *di.Container) {
	// Ping Resource
	r.HandleFunc(
		clients.ApiPingRoute,
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set(clients.ContentType, clients.ContentTypeText)
			_, _ = w.Write([]byte("pong"))
		}).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(
		clients.ApiConfigRoute,
		func(w http.ResponseWriter, _ *http.Request) {
			pkg.Encode(commandContainer.ConfigurationFrom(dic.Get), w, bootstrapContainer.LoggingClientFrom(dic.Get))
		}).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(
		clients.ApiMetricsRoute,
		func(w http.ResponseWriter, _ *http.Request) {
			pkg.Encode(telemetry.NewSystemUsage(), w, bootstrapContainer.LoggingClientFrom(dic.Get))
		}).Methods(http.MethodGet)

	// Version
	r.HandleFunc(clients.ApiVersionRoute, pkg.VersionHandler).Methods(http.MethodGet)

	b := r.PathPrefix(clients.ApiBase).Subrouter()

	loadDeviceRoutes(b, dic)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)
}

func loadDeviceRoutes(b *mux.Router, dic *di.Container) {
	b.HandleFunc(
		"/device",
		func(w http.ResponseWriter, r *http.Request) {
			restGetAllCommands(
				w,
				r,
				container.DBClientFrom(dic.Get),
				commandContainer.MetadataDeviceClientFrom(dic.Get),
				commandContainer.ConfigurationFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

	d := b.PathPrefix("/" + DEVICE).Subrouter()

	// /api/<version>/device
	d.HandleFunc(
		"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetCommandsByDeviceID(
				w,
				r,
				container.DBClientFrom(dic.Get),
				commandContainer.MetadataDeviceClientFrom(dic.Get),
				commandContainer.ConfigurationFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	d.HandleFunc(
		"/{"+ID+"}/"+COMMAND+"/{"+COMMANDID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetDeviceCommandByCommandID(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				commandContainer.MetadataDeviceClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				&http.Client{})
		}).Methods(http.MethodGet)
	d.HandleFunc(
		"/{"+ID+"}/"+COMMAND+"/{"+COMMANDID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restPutDeviceCommandByCommandID(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				commandContainer.MetadataDeviceClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				&http.Client{})
		}).Methods(http.MethodPut)
	// In the block of code above, as well as in the one that follows below,
	// there are two references each to http.Client. Putting them into the
	// DI container(dic) and retrieving the value like we do for other components
	// would bring about further consistency in the code base. But the concern
	// then would be the creation of a race condition because we can only have a
	// single http.Client instance in the dic. In turn, every invocation of this
	// REST handler would be served by a different goroutine. This would create
	// a situation where each one of them would use the same http.Client instance,
	// resulting in state divergence, misalignment. So the decision is to not
	// put this into the DI container(dic).

	// /api/<version>/device/name
	dn := d.PathPrefix("/" + NAME).Subrouter()

	dn.HandleFunc(
		"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetCommandsByDeviceName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				commandContainer.MetadataDeviceClientFrom(dic.Get),
				commandContainer.ConfigurationFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
			)
		}).Methods(http.MethodGet)
	dn.HandleFunc(
		"/{"+NAME+"}/"+COMMAND+"/{"+COMMANDNAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetDeviceCommandByNames(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				commandContainer.MetadataDeviceClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				&http.Client{})
		}).Methods(http.MethodGet)
	dn.HandleFunc(
		"/{"+NAME+"}/"+COMMAND+"/{"+COMMANDNAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restPutDeviceCommandByNames(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				commandContainer.MetadataDeviceClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				&http.Client{})
		}).Methods(http.MethodPut)
}
