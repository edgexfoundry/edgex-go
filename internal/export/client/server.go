//
// Copyright (c) 2017 Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"

	"github.com/go-zoo/bone"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
)

const (
	apiV1Registration = "/api/v1/registration"
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

// HTTPServer function
func httpServer() http.Handler {
	mux := bone.New()

	// Ping Resource
	mux.Get(clients.ApiPingRoute, http.HandlerFunc(replyPing))

	// Configuration
	mux.Get(clients.ApiConfigRoute, http.HandlerFunc(replyConfig))

	// Metrics
	mux.Get(clients.ApiMetricsRoute, http.HandlerFunc(replyMetrics))

	// Registration
	mux.Get(apiV1Registration+"/:id", http.HandlerFunc(getRegByID))
	mux.Get(apiV1Registration+"/reference/:type", http.HandlerFunc(getRegList))
	mux.Get(apiV1Registration, http.HandlerFunc(getAllReg))
	mux.Get(apiV1Registration+"/name/:name", http.HandlerFunc(getRegByName))
	mux.Post(apiV1Registration, http.HandlerFunc(addReg))
	mux.Put(apiV1Registration, http.HandlerFunc(updateReg))
	mux.Delete(apiV1Registration+"/id/:id", http.HandlerFunc(delRegByID))
	mux.Delete(apiV1Registration+"/name/:name", http.HandlerFunc(delRegByName))

	return mux
}

func StartHTTPServer(errChan chan error) {
	go func() {
		p := fmt.Sprintf(":%d", Configuration.Service.Port)
		errChan <- http.ListenAndServe(p, httpServer())
	}()
}
