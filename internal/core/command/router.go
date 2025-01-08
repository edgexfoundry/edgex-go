//
// Copyright (C) 2021-2025 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"github.com/edgexfoundry/edgex-go"
	commandController "github.com/edgexfoundry/edgex-go/internal/core/command/controller/http"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/controller"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"

	"github.com/labstack/echo/v4"
)

func LoadRestRoutes(r *echo.Echo, dic *di.Container, serviceName string) {
	authenticationHook := handlers.AutoConfigAuthenticationFunc(dic)

	// Common
	_ = controller.NewCommonController(dic, r, serviceName, edgex.Version)

	// Command
	cmd := commandController.NewCommandController(dic)
	r.GET(common.ApiAllDeviceRoute, cmd.AllCommands, authenticationHook)
	r.GET(common.ApiDeviceByNameRoute, cmd.CommandsByDeviceName, authenticationHook)
	r.GET(common.ApiDeviceNameCommandNameRoute, cmd.IssueGetCommandByName, authenticationHook)
	r.PUT(common.ApiDeviceNameCommandNameRoute, cmd.IssueSetCommandByName, authenticationHook)
}
