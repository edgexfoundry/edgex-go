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

	consulclient "github.com/edgexfoundry/consul-client-go"

	"github.com/go-zoo/bone"
	"go.uber.org/zap"
)

const (
	applicationName string = "export-client"
	consulProfile   string = "go"
)

func replyPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/text; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	str := `pong`
	io.WriteString(w, str)
}

// HTTPServer function
func httpServer() http.Handler {
	mux := bone.New()

	// Status
	mux.Get("/status", http.HandlerFunc(getStatus))

	mux.Get("/api/v1/ping", http.HandlerFunc(replyPing))

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

	// Initialize service on Consul
	err := consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    applicationName,
		ServicePort:    cfg.Port,
		ServiceAddress: "localhost",
		CheckAddress:   "http://localhost:48071/api/v1/ping",
		CheckInterval:  "10s",
		ConsulAddress:  "localhost",
		ConsulPort:     8500,
	})

	if err == nil {
		consulProfiles := []string{consulProfile}
		if err := consulclient.CheckKeyValuePairs(&cfg, applicationName, consulProfiles); err != nil {
			logger.Warn("Error getting key/values from Consul", zap.Error(err))
		}
	} else {
		logger.Warn("Error connecting to consul", zap.Error(err))
	}

	go func() {
		p := fmt.Sprintf(":%d", cfg.Port)
		logger.Info("Starting Export Client", zap.String("url", p))
		errChan <- http.ListenAndServe(p, httpServer())
	}()
}
