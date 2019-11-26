//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
)

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

	// Logs
	r.HandleFunc(clients.ApiLoggingRoute, addLog).Methods(http.MethodPost)

	r.HandleFunc(clients.ApiLoggingRoute, getLogs).Methods(http.MethodGet)
	l := r.PathPrefix(clients.ApiLoggingRoute).Subrouter()
	l.HandleFunc("/{"+LIMIT+"}", getLogs).Methods(http.MethodGet)
	l.HandleFunc("/{"+START+"}/{"+END+"}/{"+LIMIT+"}", getLogs).Methods(http.MethodGet)
	l.HandleFunc("/"+ORIGINSERVICES+"/{"+SERVICES+"}/{"+START+"}/{"+END+"}/{"+LIMIT+"}", getLogs).Methods(http.MethodGet)
	l.HandleFunc("/"+KEYWORDS+"/{"+KEYWORDS+"}/{"+START+"}/{"+END+"}/{"+LIMIT+"}", getLogs).Methods(http.MethodGet)
	l.HandleFunc("/"+LOGLEVELS+"/{"+LEVELS+"}/{"+START+"}/{"+END+"}/{"+LIMIT+"}", getLogs).Methods(http.MethodGet)
	l.HandleFunc("/"+LOGLEVELS+"/{"+LEVELS+"}/"+ORIGINSERVICES+"/{"+SERVICES+"}/{"+START+"}/{"+END+"}/{"+LIMIT+"}", getLogs).Methods(http.MethodGet)

	l.HandleFunc("/{"+START+"}/{"+END+"}", delLogs).Methods(http.MethodDelete)
	l.HandleFunc("/"+KEYWORDS+"/{"+KEYWORDS+"}/{"+START+"}/{"+END+"}", delLogs).Methods(http.MethodDelete)
	l.HandleFunc("/"+ORIGINSERVICES+"/{"+SERVICES+"}/{"+START+"}/{"+END+"}", delLogs).Methods(http.MethodDelete)
	l.HandleFunc("/"+LOGLEVELS+"/{"+LEVELS+"}/{"+START+"}/{"+END+"}", delLogs).Methods(http.MethodDelete)
	l.HandleFunc("/"+LOGLEVELS+"/{"+LEVELS+"}/"+ORIGINSERVICES+"/{"+SERVICES+"}/{"+START+"}/{"+END+"}", delLogs).Methods(http.MethodDelete)
	l.HandleFunc("/"+REMOVEOLD+"/"+AGE+"/{"+AGE+"}", delLogs).Methods(http.MethodDelete)

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

	return
}
