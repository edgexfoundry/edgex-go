//
// Copyright (c) 2017 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"

	"github.com/go-zoo/bone"
	"github.com/edgexfoundry/edgex-go/internal/export"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/edgexfoundry/edgex-go/internal"
)

func replyPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	str := `pong`
	io.WriteString(w, str)
}

func replyConfig(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	encode(Configuration, w)
}

func replyNotifyRegistrations(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed read body. Error: %s", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	update := models.NotifyUpdate{}
	if err := json.Unmarshal(data, &update); err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to parse %X", data))
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}
	if update.Name == "" || update.Operation == "" {
		LoggingClient.Error(fmt.Sprintf("Missing json field: %s", update.Name))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if update.Operation != export.NotifyUpdateAdd &&
		update.Operation != export.NotifyUpdateUpdate &&
		update.Operation != export.NotifyUpdateDelete {
		LoggingClient.Error(fmt.Sprintf("Invalid value for operation %s", update.Operation))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	RefreshRegistrations(update)
}

func replyMetrics(w http.ResponseWriter, r *http.Request) {

	var t internal.Telemetry

	if r.Body != nil {
		defer r.Body.Close()
	}

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

// HTTPServer function
func httpServer() http.Handler {
	mux := bone.New()

	// Ping Resource
	mux.Get(clients.ApiPingRoute, http.HandlerFunc(replyPing))

	// Configuration
	mux.Get(clients.ApiConfigRoute, http.HandlerFunc(replyConfig))

	// Metrics
	mux.Get(clients.ApiMetricsRoute, http.HandlerFunc(replyMetrics))

	mux.Put(clients.ApiNotifyRegistrationRoute, http.HandlerFunc(replyNotifyRegistrations))

	return mux
}
