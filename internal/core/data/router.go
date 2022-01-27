//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"

	dataController "github.com/edgexfoundry/edgex-go/internal/core/data/controller/http"
	commonController "github.com/edgexfoundry/edgex-go/internal/pkg/controller/http"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
)

func LoadRestRoutes(r *mux.Router, dic *di.Container, serviceName string) {
	// Common
	cc := commonController.NewCommonController(dic, serviceName)
	r.HandleFunc(common.ApiPingRoute, cc.Ping).Methods(http.MethodGet)
	r.HandleFunc(common.ApiVersionRoute, cc.Version).Methods(http.MethodGet)
	r.HandleFunc(common.ApiConfigRoute, cc.Config).Methods(http.MethodGet)
	r.HandleFunc(common.ApiMetricsRoute, cc.Metrics).Methods(http.MethodGet)

	// Events
	ec := dataController.NewEventController(dic)
	r.HandleFunc(common.ApiEventProfileNameDeviceNameSourceNameRoute, ec.AddEvent).Methods(http.MethodPost)
	r.HandleFunc(common.ApiEventIdRoute, ec.EventById).Methods(http.MethodGet)
	r.HandleFunc(common.ApiEventIdRoute, ec.DeleteEventById).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiEventCountRoute, ec.EventTotalCount).Methods(http.MethodGet)
	r.HandleFunc(common.ApiEventCountByDeviceNameRoute, ec.EventCountByDeviceName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiAllEventRoute, ec.AllEvents).Methods(http.MethodGet)
	r.HandleFunc(common.ApiEventByDeviceNameRoute, ec.EventsByDeviceName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiEventByDeviceNameRoute, ec.DeleteEventsByDeviceName).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiEventByTimeRangeRoute, ec.EventsByTimeRange).Methods(http.MethodGet)
	r.HandleFunc(common.ApiEventByAgeRoute, ec.DeleteEventsByAge).Methods(http.MethodDelete)

	// Readings
	rc := dataController.NewReadingController(dic)
	r.HandleFunc(common.ApiReadingCountRoute, rc.ReadingTotalCount).Methods(http.MethodGet)
	r.HandleFunc(common.ApiAllReadingRoute, rc.AllReadings).Methods(http.MethodGet)
	r.HandleFunc(common.ApiReadingByDeviceNameRoute, rc.ReadingsByDeviceName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiReadingByTimeRangeRoute, rc.ReadingsByTimeRange).Methods(http.MethodGet)
	r.HandleFunc(common.ApiReadingByResourceNameRoute, rc.ReadingsByResourceName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiReadingCountByDeviceNameRoute, rc.ReadingCountByDeviceName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiReadingByResourceNameAndTimeRangeRoute, rc.ReadingsByResourceNameAndTimeRange).Methods(http.MethodGet)
	r.HandleFunc(common.ApiReadingByDeviceNameAndResourceNameRoute, rc.ReadingsByDeviceNameAndResourceName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiReadingByDeviceNameAndResourceNameAndTimeRangeRoute, rc.ReadingsByDeviceNameAndResourceNameAndTimeRange).Methods(http.MethodGet)
	r.HandleFunc(common.ApiReadingByDeviceNameAndTimeRangeRoute, rc.ReadingsByDeviceNameAndResourceNamesAndTimeRange).Methods(http.MethodGet)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.LoggingMiddleware(container.LoggingClientFrom(dic.Get)))
}
