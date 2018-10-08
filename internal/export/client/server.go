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

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/go-zoo/bone"
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

	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	err := enc.Encode(Configuration)
	// Problems encoding
	if err != nil {
		LoggingClient.Error("Error encoding the data: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HTTPServer function
func httpServer() http.Handler {
	mux := bone.New()

	mux.Get(internal.ApiPingRoute, http.HandlerFunc(replyPing))
	mux.Get(internal.ApiConfigRoute, http.HandlerFunc(replyConfig))

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
