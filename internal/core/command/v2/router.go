//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package v2

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/gorilla/mux"

	commandController "github.com/edgexfoundry/edgex-go/internal/core/command/v2/controller/http"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	commonController "github.com/edgexfoundry/edgex-go/internal/pkg/v2/controller/http"
)

func LoadRestRoutes(r *mux.Router, dic *di.Container) {
	// v2 API routes
	// Common
	cc := commonController.NewV2CommonController(dic)
	r.HandleFunc(v2.ApiPingRoute, cc.Ping).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiVersionRoute, cc.Version).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiConfigRoute, cc.Config).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiMetricsRoute, cc.Metrics).Methods(http.MethodGet)

	// Command
	cmd := commandController.NewCommandController(dic)
	r.HandleFunc(v2.ApiAllDeviceRoute, cmd.AllCommands).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiDeviceByNameRoute, cmd.CommandsByDeviceName).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiDeviceNameCommandNameRoute, cmd.IssueGetCommandByName).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiDeviceNameCommandNameRoute, cmd.IssueSetCommandByName).Methods(http.MethodPut)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.LoggingMiddleware(container.LoggingClientFrom(dic.Get)))
}
