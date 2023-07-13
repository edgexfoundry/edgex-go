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

	g := r.Group(common.ApiBase)
	g.Use(authenticationHook)
	g.Use(correlation.ManageHeader)
	g.Use(correlation.LoggingMiddleware(container.LoggingClientFrom(dic.Get)))
	g.Use(correlation.UrlDecodeMiddleware(container.LoggingClientFrom(dic.Get)))

	// Command
	cmd := commandController.NewCommandController(dic)

	g.GET("/"+common.Device+"/"+common.All, cmd.AllCommands)
	g.GET("/"+common.Device+"/"+common.Name+"/:"+common.Name, cmd.CommandsByDeviceName)
	g.GET("/"+common.Device+"/"+common.Name+"/:"+common.Name+"/:"+common.Command, cmd.IssueGetCommandByName)
	g.PUT("/"+common.Device+"/"+common.Name+"/:"+common.Name+"/:"+common.Command, cmd.IssueSetCommandByName)
}
