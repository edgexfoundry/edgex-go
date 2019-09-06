//
// Copyright (c) 2017 Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/gorilla/mux"
)

// Test if the service is working
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong"))
}

func configHandler(w http.ResponseWriter, _ *http.Request) {
	pkg.Encode(Configuration, w, LoggingClient)
}

func metricsHandler(w http.ResponseWriter, _ *http.Request) {
	s := telemetry.NewSystemUsage()

	pkg.Encode(s, w, LoggingClient)

	return
}

func LoadRestRoutes() *mux.Router {
	r := mux.NewRouter()

	// Ping Resource
	r.HandleFunc(clients.ApiPingRoute, pingHandler).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(clients.ApiConfigRoute, configHandler).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(clients.ApiMetricsRoute, metricsHandler).Methods(http.MethodGet)

	// Version
	r.HandleFunc(clients.ApiVersionRoute, pkg.VersionHandler).Methods(http.MethodGet)

	// Registration
	r.HandleFunc(clients.ApiRegistrationRoute, getAllReg).Methods(http.MethodGet)
	r.HandleFunc(clients.ApiRegistrationRoute, addReg).Methods(http.MethodPost)
	r.HandleFunc(clients.ApiRegistrationRoute, updateReg).Methods(http.MethodPut)
	reg := r.PathPrefix(clients.ApiRegistrationRoute).Subrouter()
	reg.HandleFunc("/{"+ID+"}", getRegByID).Methods(http.MethodGet)
	reg.HandleFunc("/"+REFERENCE+"/{"+TYPE+"}", getRegList).Methods(http.MethodGet)
	reg.HandleFunc("/"+NAME+"/{"+NAME+"}", getRegByName).Methods(http.MethodGet)
	reg.HandleFunc("/"+ID+"/{"+ID+"}", delRegByID).Methods(http.MethodDelete)
	reg.HandleFunc("/"+NAME+"/{"+NAME+"}", delRegByName).Methods(http.MethodDelete)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)

	return r
}
