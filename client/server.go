//
// Copyright (c) 2017 Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"fmt"
	"net/http"

	"github.com/go-zoo/bone"
	"go.uber.org/zap"
)

// HTTPServer function
func httpServer() http.Handler {
	mux := bone.New()

	// Status
	mux.Get("/status", http.HandlerFunc(getStatus))

	// Registration
	mux.Get("/api/v1/registration/:id", http.HandlerFunc(getRegByID))
	mux.Get("/api/v1/registration/reference/:type", http.HandlerFunc(getRegList))
	mux.Get("/api/v1/registration", http.HandlerFunc(getAllReg))
	mux.Get("/api/v1/registration/name/:name", http.HandlerFunc(getRegByName))
	mux.Post("/api/v1/registration", http.HandlerFunc(addReg))
	mux.Put("/api/v1/registration", http.HandlerFunc(updateReg))
	mux.Delete("/api/v1/registration/id/:id", http.HandlerFunc(delRegByID))
	mux.Delete("/api/v1/registration/name/:name", http.HandlerFunc(delRegByName))

	return mux
}

func StartHTTPServer(config Config, errChan chan error) {
	cfg = config
	go func() {
		p := fmt.Sprintf(":%d", cfg.Port)
		logger.Info("Starting Export Client", zap.String("url", p))
		errChan <- http.ListenAndServe(p, httpServer())
	}()
}
