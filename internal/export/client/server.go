//
// Copyright (c) 2017 Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"fmt"
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
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

// HTTPServer function
func httpServer() http.Handler {
	r := mux.NewRouter()

	// Ping Resource
	r.HandleFunc(clients.ApiPingRoute, pingHandler).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(clients.ApiConfigRoute, configHandler).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(clients.ApiMetricsRoute, metricsHandler).Methods(http.MethodGet)

	// Registration
	r.HandleFunc(clients.ApiRegistrationRoute, getAllReg).Methods(http.MethodGet)
	r.HandleFunc(clients.ApiRegistrationRoute, addReg).Methods(http.MethodPost)
	r.HandleFunc(clients.ApiRegistrationRoute, updateReg).Methods(http.MethodPut)
	reg := r.PathPrefix(clients.ApiRegistrationRoute).Subrouter()
	reg.HandleFunc("/{id}", getRegByID).Methods(http.MethodGet)
	reg.HandleFunc("/reference/{type}", getRegList).Methods(http.MethodGet)
	reg.HandleFunc("/name/{name}", getRegByName).Methods(http.MethodGet)
	reg.HandleFunc("/id/{id}", delRegByID).Methods(http.MethodDelete)
	reg.HandleFunc("/name/{name}", delRegByName).Methods(http.MethodDelete)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)

	return r
}

func StartHTTPServer(errChan chan error) {
	go func() {
		p := fmt.Sprintf(":%d", Configuration.Service.Port)
		errChan <- http.ListenAndServe(p, httpServer())
	}()
}
