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

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
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
	s := telemetry.NewSystemUsage()

	encode(s, w)

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

func StartHTTPServer(errChan chan error) {
	go func() {
		p := fmt.Sprintf(":%d", Configuration.Service.Port)
		errChan <- http.ListenAndServe(p, httpServer())
	}()
}
