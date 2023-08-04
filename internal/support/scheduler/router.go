// Copyright (C) 2021 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/controller"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"

	schedulerController "github.com/edgexfoundry/edgex-go/internal/support/scheduler/controller/http"

	"github.com/labstack/echo/v4"
)

func LoadRestRoutes(r *echo.Echo, dic *di.Container, serviceName string) {
	lc := container.LoggingClientFrom(dic.Get)
	secretProvider := container.SecretProviderExtFrom(dic.Get)
	authenticationHook := handlers.AutoConfigAuthenticationFunc(secretProvider, lc)

	// Common
	_ = controller.NewCommonController(dic, r, serviceName, edgex.Version)

	// Interval
	interval := schedulerController.NewIntervalController(dic)
	r.POST(common.ApiIntervalRoute, interval.AddInterval, authenticationHook)
	r.GET(common.ApiIntervalByNameEchoRoute, interval.IntervalByName, authenticationHook)
	r.GET(common.ApiAllIntervalRoute, interval.AllIntervals, authenticationHook)
	r.DELETE(common.ApiIntervalByNameEchoRoute, interval.DeleteIntervalByName, authenticationHook)
	r.PATCH(common.ApiIntervalRoute, interval.PatchInterval, authenticationHook)

	// IntervalAction
	action := schedulerController.NewIntervalActionController(dic)
	r.POST(common.ApiIntervalActionRoute, action.AddIntervalAction, authenticationHook)
	r.GET(common.ApiAllIntervalActionRoute, action.AllIntervalActions, authenticationHook)
	r.GET(common.ApiIntervalActionByNameEchoRoute, action.IntervalActionByName, authenticationHook)
	r.DELETE(common.ApiIntervalActionByNameEchoRoute, action.DeleteIntervalActionByName, authenticationHook)
	r.PATCH(common.ApiIntervalActionRoute, action.PatchIntervalAction, authenticationHook)
}
