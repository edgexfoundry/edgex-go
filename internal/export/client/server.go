//
// Copyright (c) 2017 Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-zoo/bone"

)

const (
	apiV1Registration = "/api/v1/registration"
	apiV1Ping         = "/api/v1/ping"
)

func replyPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	str := `pong`
	io.WriteString(w, str)
}

// HTTPServer function
func httpServer() http.Handler {
	mux := bone.New()

	mux.Get(apiV1Ping, http.HandlerFunc(replyPing))

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

func StartHTTPServer(config ConfigurationStruct, errChan chan error) {
	go func() {
		p := fmt.Sprintf(":%d", config.Port)
		logger.Info("Starting Export Client", logger.String("url", p))
		errChan <- http.ListenAndServe(p, httpServer())
	}()
}
