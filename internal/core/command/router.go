//
// Copyright (C) 2021-2023 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/gorilla/mux"

	commandController "github.com/edgexfoundry/edgex-go/internal/core/command/controller/http"
	commonController "github.com/edgexfoundry/edgex-go/internal/pkg/controller/http"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
)

func LoadRestRoutes(r *mux.Router, dic *di.Container, serviceName string) {
	// r.UseEncodedPath() tells the router to match the encoded original path to the routes
	r.UseEncodedPath()

	lc := container.LoggingClientFrom(dic.Get)
	secretProvider := container.SecretProviderExtFrom(dic.Get)
	authenticationHook := handlers.AutoConfigAuthenticationFunc(secretProvider, lc)

	// Common
	cc := commonController.NewCommonController(dic, serviceName)
	r.HandleFunc(common.ApiPingRoute, cc.Ping).Methods(http.MethodGet) // Health check is always unauthenticated
	r.HandleFunc(common.ApiVersionRoute, authenticationHook(cc.Version)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiConfigRoute, authenticationHook(cc.Config)).Methods(http.MethodGet)

	// Command
	cmd := commandController.NewCommandController(dic)
	r.HandleFunc(common.ApiAllDeviceRoute, authenticationHook(cmd.AllCommands)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceByNameRoute, authenticationHook(cmd.CommandsByDeviceName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceNameCommandNameRoute, authenticationHook(cmd.IssueGetCommandByName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceNameCommandNameRoute, authenticationHook(cmd.IssueSetCommandByName)).Methods(http.MethodPut)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.LoggingMiddleware(container.LoggingClientFrom(dic.Get)))
	r.Use(correlation.UrlDecodeMiddleware(container.LoggingClientFrom(dic.Get)))
}
