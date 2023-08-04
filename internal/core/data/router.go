//
// Copyright (C) 2021-2023 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/controller"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"

	dataController "github.com/edgexfoundry/edgex-go/internal/core/data/controller/http"

	"github.com/labstack/echo/v4"
)

func LoadRestRoutes(r *echo.Echo, dic *di.Container, serviceName string) {
	lc := container.LoggingClientFrom(dic.Get)
	secretProvider := container.SecretProviderExtFrom(dic.Get)
	authenticationHook := handlers.AutoConfigAuthenticationFunc(secretProvider, lc)

	// Common
	_ = controller.NewCommonController(dic, r, serviceName, edgex.Version)

	// Events
	ec := dataController.NewEventController(dic)
	r.POST(common.ApiEventServiceNameProfileNameDeviceNameSourceNameEchoRoute, ec.AddEvent, authenticationHook)
	r.GET(common.ApiEventIdEchoRoute, ec.EventById, authenticationHook)
	r.DELETE(common.ApiEventIdEchoRoute, ec.DeleteEventById, authenticationHook)
	r.GET(common.ApiEventCountRoute, ec.EventTotalCount, authenticationHook)
	r.GET(common.ApiEventCountByDeviceNameEchoRoute, ec.EventCountByDeviceName, authenticationHook)
	r.GET(common.ApiAllEventRoute, ec.AllEvents, authenticationHook)
	r.GET(common.ApiEventByDeviceNameEchoRoute, ec.EventsByDeviceName, authenticationHook)
	r.DELETE(common.ApiEventByDeviceNameEchoRoute, ec.DeleteEventsByDeviceName, authenticationHook)
	r.GET(common.ApiEventByTimeRangeEchoRoute, ec.EventsByTimeRange, authenticationHook)
	r.DELETE(common.ApiEventByAgeEchoRoute, ec.DeleteEventsByAge, authenticationHook) // TODO: Add authentication to support-scheduler

	// Readings
	rc := dataController.NewReadingController(dic)
	r.GET(common.ApiReadingCountRoute, rc.ReadingTotalCount, authenticationHook)
	r.GET(common.ApiAllReadingRoute, rc.AllReadings, authenticationHook)
	r.GET(common.ApiReadingByDeviceNameEchoRoute, rc.ReadingsByDeviceName, authenticationHook)
	r.GET(common.ApiReadingByTimeRangeEchoRoute, rc.ReadingsByTimeRange, authenticationHook)
	r.GET(common.ApiReadingByResourceNameEchoRoute, rc.ReadingsByResourceName, authenticationHook)
	r.GET(common.ApiReadingCountByDeviceNameEchoRoute, rc.ReadingCountByDeviceName, authenticationHook)
	r.GET(common.ApiReadingByResourceNameAndTimeRangeEchoRoute, rc.ReadingsByResourceNameAndTimeRange, authenticationHook)
	r.GET(common.ApiReadingByDeviceNameAndResourceNameEchoRoute, rc.ReadingsByDeviceNameAndResourceName, authenticationHook)
	r.GET(common.ApiReadingByDeviceNameAndResourceNameAndTimeRangeEchoRoute, rc.ReadingsByDeviceNameAndResourceNameAndTimeRange, authenticationHook)
	r.GET(common.ApiReadingByDeviceNameAndTimeRangeEchoRoute, rc.ReadingsByDeviceNameAndResourceNamesAndTimeRange, authenticationHook)
}
