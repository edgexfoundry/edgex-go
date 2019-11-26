//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/container"
)

func LoadRestRoutes(dic *di.Container) *mux.Router {
	r := mux.NewRouter()

	// Ping Resource
	r.HandleFunc(clients.ApiPingRoute, pingHandler).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(clients.ApiConfigRoute, configHandler).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(clients.ApiMetricsRoute, metricsHandler).Methods(http.MethodGet)

	// Version
	r.HandleFunc(clients.ApiVersionRoute, pkg.VersionHandler).Methods(http.MethodGet)

	// Logs
	r.HandleFunc(clients.ApiLoggingRoute, func(w http.ResponseWriter, r *http.Request) {
		addLog(
			w,
			r,
			container.PersistenceFrom(dic.Get))
	}).Methods(http.MethodPost)

	r.HandleFunc(clients.ApiLoggingRoute, func(w http.ResponseWriter, r *http.Request) {
		getLogs(
			w,
			r,
			container.PersistenceFrom(dic.Get))
	}).Methods(http.MethodGet)

	l := r.PathPrefix(clients.ApiLoggingRoute).Subrouter()
	l.HandleFunc("/{"+LIMIT+"}", func(w http.ResponseWriter, r *http.Request) {
		getLogs(
			w,
			r,
			container.PersistenceFrom(dic.Get))
	}).Methods(http.MethodGet)
	l.HandleFunc("/{"+START+"}/{"+END+"}/{"+LIMIT+"}", func(w http.ResponseWriter, r *http.Request) {
		getLogs(
			w,
			r,
			container.PersistenceFrom(dic.Get))
	}).Methods(http.MethodGet)
	l.HandleFunc("/"+ORIGINSERVICES+"/{"+SERVICES+"}/{"+START+"}/{"+END+"}/{"+LIMIT+"}",
		func(w http.ResponseWriter, r *http.Request) {
			getLogs(
				w,
				r,
				container.PersistenceFrom(dic.Get))
		}).Methods(http.MethodGet)
	l.HandleFunc("/"+KEYWORDS+"/{"+KEYWORDS+"}/{"+START+"}/{"+END+"}/{"+LIMIT+"}",
		func(w http.ResponseWriter, r *http.Request) {
			getLogs(
				w,
				r,
				container.PersistenceFrom(dic.Get))
		}).Methods(http.MethodGet)
	l.HandleFunc("/"+LOGLEVELS+"/{"+LEVELS+"}/{"+START+"}/{"+END+"}/{"+LIMIT+"}",
		func(w http.ResponseWriter, r *http.Request) {
			getLogs(
				w,
				r,
				container.PersistenceFrom(dic.Get))
		}).Methods(http.MethodGet)
	l.HandleFunc("/"+LOGLEVELS+"/{"+LEVELS+"}/"+ORIGINSERVICES+"/{"+SERVICES+"}/{"+START+"}/{"+END+"}/{"+LIMIT+"}",
		func(w http.ResponseWriter, r *http.Request) {
			getLogs(
				w,
				r,
				container.PersistenceFrom(dic.Get))
		}).Methods(http.MethodGet)

	l.HandleFunc("/{"+START+"}/{"+END+"}", func(w http.ResponseWriter, r *http.Request) {
		delLogs(
			w,
			r,
			container.PersistenceFrom(dic.Get))
	}).Methods(http.MethodDelete)
	l.HandleFunc("/"+KEYWORDS+"/{"+KEYWORDS+"}/{"+START+"}/{"+END+"}",
		func(w http.ResponseWriter, r *http.Request) {
			delLogs(
				w,
				r,
				container.PersistenceFrom(dic.Get))
		}).Methods(http.MethodDelete)
	l.HandleFunc("/"+ORIGINSERVICES+"/{"+SERVICES+"}/{"+START+"}/{"+END+"}",
		func(w http.ResponseWriter, r *http.Request) {
			delLogs(
				w,
				r,
				container.PersistenceFrom(dic.Get))
		}).Methods(http.MethodDelete)
	l.HandleFunc("/"+LOGLEVELS+"/{"+LEVELS+"}/{"+START+"}/{"+END+"}", func(w http.ResponseWriter, r *http.Request) {
		delLogs(
			w,
			r,
			container.PersistenceFrom(dic.Get))
	}).Methods(http.MethodDelete)
	l.HandleFunc("/"+LOGLEVELS+"/{"+LEVELS+"}/"+ORIGINSERVICES+"/{"+SERVICES+"}/{"+START+"}/{"+END+"}",
		func(w http.ResponseWriter, r *http.Request) {
			delLogs(
				w,
				r,
				container.PersistenceFrom(dic.Get))
		}).Methods(http.MethodDelete)
	l.HandleFunc("/"+REMOVEOLD+"/"+AGE+"/{"+AGE+"}", func(w http.ResponseWriter, r *http.Request) {
		delLogs(
			w,
			r,
			container.PersistenceFrom(dic.Get))
	}).Methods(http.MethodDelete)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)

	return r
}

// Test if the service is working
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(clients.ContentType, clients.ContentTypeText)
	w.Write([]byte("pong"))
}

func configHandler(w http.ResponseWriter, _ *http.Request) {
	pkg.Encode(Configuration, w, LoggingClient)
}

func metricsHandler(w http.ResponseWriter, _ *http.Request) {
	s := telemetry.NewSystemUsage()

	pkg.Encode(s, w, LoggingClient)
}
