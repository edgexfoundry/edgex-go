//
// Copyright (c) 2017 Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/gorilla/mux"
)

// Test if the service is working
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong"))
}

func configHandler(w http.ResponseWriter, _ *http.Request) {
	encode(Configuration, w)
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

func metricsHandler(w http.ResponseWriter, _ *http.Request) {
	var t internal.Telemetry

	// The micro-service is to be considered the System Of Record (SOR) in terms of accurate information.
	// Fetch metrics for the scheduler service.
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

	return r
}

func StartHTTPServer(errChan chan error) {
	go func() {
		p := fmt.Sprintf(":%d", Configuration.Service.Port)
		errChan <- http.ListenAndServe(p, httpServer())
	}()
}
