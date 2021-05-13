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

	dataController "github.com/edgexfoundry/edgex-go/internal/core/data/v2/controller/http"
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

	// Events
	ec := dataController.NewEventController(dic)
	r.HandleFunc(v2.ApiEventProfileNameDeviceNameSourceNameRoute, ec.AddEvent).Methods(http.MethodPost)
	r.HandleFunc(v2.ApiEventIdRoute, ec.EventById).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiEventIdRoute, ec.DeleteEventById).Methods(http.MethodDelete)
	r.HandleFunc(v2.ApiEventCountRoute, ec.EventTotalCount).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiEventCountByDeviceNameRoute, ec.EventCountByDeviceName).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiAllEventRoute, ec.AllEvents).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiEventByDeviceNameRoute, ec.EventsByDeviceName).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiEventByDeviceNameRoute, ec.DeleteEventsByDeviceName).Methods(http.MethodDelete)
	r.HandleFunc(v2.ApiEventByTimeRangeRoute, ec.EventsByTimeRange).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiEventByAgeRoute, ec.DeleteEventsByAge).Methods(http.MethodDelete)

	// Readings
	rc := dataController.NewReadingController(dic)
	r.HandleFunc(v2.ApiReadingCountRoute, rc.ReadingTotalCount).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiAllReadingRoute, rc.AllReadings).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiReadingByDeviceNameRoute, rc.ReadingsByDeviceName).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiReadingByTimeRangeRoute, rc.ReadingsByTimeRange).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiReadingByResourceNameRoute, rc.ReadingsByResourceName).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiReadingCountByDeviceNameRoute, rc.ReadingCountByDeviceName).Methods(http.MethodGet)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.LoggingMiddleware(container.LoggingClientFrom(dic.Get)))
}
