/*******************************************************************************
 * Copyright 2018 Dell Inc.
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

package scheduler

import (
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
	schedulerContainer "github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"

	"github.com/gorilla/mux"
)

func loadRestRoutes(r *mux.Router, dic *di.Container) {
	// Ping Resource
	r.HandleFunc(clients.
		ApiPingRoute,
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set(clients.ContentType, clients.ContentTypeText)
			_, _ = w.Write([]byte("pong"))
		}).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(clients.
		ApiConfigRoute,
		func(w http.ResponseWriter, _ *http.Request) {
			pkg.Encode(schedulerContainer.ConfigurationFrom(dic.Get), w, bootstrapContainer.LoggingClientFrom(dic.Get))
		}).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(clients.
		ApiMetricsRoute,
		func(w http.ResponseWriter, _ *http.Request) {
			pkg.Encode(telemetry.NewSystemUsage(), w, bootstrapContainer.LoggingClientFrom(dic.Get))
		}).Methods(http.MethodGet)

	// Version
	r.HandleFunc(clients.ApiVersionRoute, pkg.VersionHandler).Methods(http.MethodGet)

	// Interval
	r.HandleFunc(clients.
		ApiIntervalRoute,
		func(w http.ResponseWriter, r *http.Request) {
			restGetIntervals(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				schedulerContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)
	r.HandleFunc(clients.
		ApiIntervalRoute,
		func(w http.ResponseWriter, r *http.Request) {
			restUpdateInterval(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				schedulerContainer.QueueFrom(dic.Get))
		}).Methods(http.MethodPut)
	r.HandleFunc(clients.
		ApiIntervalRoute,
		func(w http.ResponseWriter, r *http.Request) {
			restAddInterval(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				schedulerContainer.QueueFrom(dic.Get))
		}).Methods(http.MethodPost)
	interval := r.PathPrefix(clients.ApiIntervalRoute).Subrouter()
	interval.HandleFunc(
		"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetIntervalByID(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	interval.HandleFunc(
		"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteIntervalByID(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				schedulerContainer.QueueFrom(dic.Get))
		}).Methods(http.MethodDelete)
	interval.HandleFunc(
		"/"+NAME+"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetIntervalByName(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	interval.HandleFunc(
		"/"+NAME+"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteIntervalByName(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				schedulerContainer.QueueFrom(dic.Get))
		}).Methods(http.MethodDelete)
	// Scrub "Intervals and IntervalActions"
	interval.HandleFunc(
		"/"+SCRUB+"/",
		func(w http.ResponseWriter, r *http.Request) {
			restScrubAllIntervals(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)

	// IntervalAction
	r.HandleFunc(clients.
		ApiIntervalActionRoute,
		func(w http.ResponseWriter, r *http.Request) {
			restGetIntervalAction(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				schedulerContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)
	r.HandleFunc(clients.
		ApiIntervalActionRoute,
		func(w http.ResponseWriter, r *http.Request) {
			restAddIntervalAction(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				schedulerContainer.QueueFrom(dic.Get))
		}).Methods(http.MethodPost)
	r.HandleFunc(clients.
		ApiIntervalActionRoute,
		func(w http.ResponseWriter, r *http.Request) {
			intervalActionHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				schedulerContainer.QueueFrom(dic.Get),
				schedulerContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPut)
	intervalAction := r.PathPrefix(clients.ApiIntervalActionRoute).Subrouter()
	intervalAction.HandleFunc(
		"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			intervalActionByIdHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				schedulerContainer.QueueFrom(dic.Get))
		}).Methods(http.MethodGet, http.MethodDelete)
	intervalAction.HandleFunc(
		"/"+NAME+"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			intervalActionByNameHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				schedulerContainer.QueueFrom(dic.Get))
		}).Methods(http.MethodGet, http.MethodDelete)
	intervalAction.HandleFunc(
		"/"+TARGET+"/{"+TARGET+"}",
		func(w http.ResponseWriter, r *http.Request) {
			intervalActionByTargetHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	intervalAction.HandleFunc(
		"/"+INTERVAL+"/{"+INTERVAL+"}",
		func(w http.ResponseWriter, r *http.Request) {
			intervalActionByIntervalHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)

	// Scrub "IntervalActions"
	intervalAction.HandleFunc(
		"/"+SCRUB+"/",
		func(w http.ResponseWriter, r *http.Request) {
			scrubIntervalActionsHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)
}
