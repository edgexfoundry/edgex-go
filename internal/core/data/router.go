//
// Copyright (C) 2021-2025 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/controller"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"

	dataController "github.com/edgexfoundry/edgex-go/internal/core/data/controller/http"

	"github.com/labstack/echo/v4"
)

func LoadRestRoutes(r *echo.Echo, dic *di.Container, serviceName string) {
	authenticationHook := handlers.AutoConfigAuthenticationFunc(dic)

	// Common
	_ = controller.NewCommonController(dic, r, serviceName, edgex.Version)

	// Events
	ec := dataController.NewEventController(dic)
	r.POST(common.ApiEventServiceNameProfileNameDeviceNameSourceNameRoute, ec.AddEvent, authenticationHook)
	r.GET(common.ApiEventIdRoute, ec.EventById, authenticationHook)
	r.DELETE(common.ApiEventIdRoute, ec.DeleteEventById, authenticationHook)
	r.GET(common.ApiEventCountRoute, ec.EventTotalCount, authenticationHook)
	r.GET(common.ApiEventCountByDeviceNameRoute, ec.EventCountByDeviceName, authenticationHook)
	r.GET(common.ApiAllEventRoute, ec.AllEvents, authenticationHook)
	r.GET(common.ApiEventByDeviceNameRoute, ec.EventsByDeviceName, authenticationHook)
	r.DELETE(common.ApiEventByDeviceNameRoute, ec.DeleteEventsByDeviceName, authenticationHook)
	r.GET(common.ApiEventByTimeRangeRoute, ec.EventsByTimeRange, authenticationHook)
	r.DELETE(common.ApiEventByAgeRoute, ec.DeleteEventsByAge, authenticationHook) // TODO: Add authentication to support-scheduler

	// Readings
	rc := dataController.NewReadingController(dic)
	r.GET(common.ApiReadingCountRoute, rc.ReadingTotalCount, authenticationHook)
	r.GET(common.ApiAllReadingRoute, rc.AllReadings, authenticationHook)
	r.GET(common.ApiReadingByDeviceNameRoute, rc.ReadingsByDeviceName, authenticationHook)
	r.GET(common.ApiReadingByTimeRangeRoute, rc.ReadingsByTimeRange, authenticationHook)
	r.GET(common.ApiReadingByResourceNameRoute, rc.ReadingsByResourceName, authenticationHook)
	r.GET(common.ApiReadingCountByDeviceNameRoute, rc.ReadingCountByDeviceName, authenticationHook)
	r.GET(common.ApiReadingByResourceNameAndTimeRangeRoute, rc.ReadingsByResourceNameAndTimeRange, authenticationHook)
	r.GET(common.ApiReadingByDeviceNameAndResourceNameRoute, rc.ReadingsByDeviceNameAndResourceName, authenticationHook)
	r.GET(common.ApiReadingByDeviceNameAndResourceNameAndTimeRangeRoute, rc.ReadingsByDeviceNameAndResourceNameAndTimeRange, authenticationHook)
	r.GET(common.ApiReadingByDeviceNameAndTimeRangeRoute, rc.ReadingsByDeviceNameAndResourceNamesAndTimeRange, authenticationHook)
}
