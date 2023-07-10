//
// Copyright (C) 2021-2023 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"github.com/edgexfoundry/edgex-go"
	commandController "github.com/edgexfoundry/edgex-go/internal/core/command/controller/http"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/controller"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"

	"github.com/labstack/echo/v4"
)

func LoadRestRoutes(r *echo.Echo, dic *di.Container, serviceName string) {
	lc := container.LoggingClientFrom(dic.Get)
	secretProvider := container.SecretProviderExtFrom(dic.Get)
	authenticationHook := handlers.AutoConfigAuthenticationFunc(secretProvider, lc)

	// Common
	_ = controller.NewCommonController(dic, r, serviceName, edgex.Version)

	// Command
	cmd := commandController.NewCommandController(dic)

	// create a route group with /api/v3/device as prefix, which applies the same authenticationHook middleware
	deviceRoutes := r.Group(common.ApiDeviceRoute)
	deviceRoutes.Use(authenticationHook)

	deviceRoutes.GET("/"+common.All, cmd.AllCommands)
	deviceRoutes.GET("/"+common.Name+"/:"+common.Name, cmd.CommandsByDeviceName)
	deviceRoutes.GET("/"+common.Name+"/:"+common.Name+"/:"+common.Command, cmd.IssueGetCommandByName)
	deviceRoutes.PUT("/"+common.Name+"/:"+common.Name+"/:"+common.Command, cmd.IssueSetCommandByName)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.LoggingMiddleware(container.LoggingClientFrom(dic.Get)))
}
