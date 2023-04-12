//
// Copyright (C) 2021-2023 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"

	dataController "github.com/edgexfoundry/edgex-go/internal/core/data/controller/http"
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

	// Events
	ec := dataController.NewEventController(dic)
	r.HandleFunc(common.ApiEventServiceNameProfileNameDeviceNameSourceNameRoute, authenticationHook(ec.AddEvent)).Methods(http.MethodPost)
	r.HandleFunc(common.ApiEventIdRoute, authenticationHook(ec.EventById)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiEventIdRoute, authenticationHook(ec.DeleteEventById)).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiEventCountRoute, authenticationHook(ec.EventTotalCount)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiEventCountByDeviceNameRoute, authenticationHook(ec.EventCountByDeviceName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiAllEventRoute, authenticationHook(ec.AllEvents)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiEventByDeviceNameRoute, authenticationHook(ec.EventsByDeviceName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiEventByDeviceNameRoute, authenticationHook(ec.DeleteEventsByDeviceName)).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiEventByTimeRangeRoute, authenticationHook(ec.EventsByTimeRange)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiEventByAgeRoute, authenticationHook(ec.DeleteEventsByAge)).Methods(http.MethodDelete) // TODO: Add authentication to support-scheduler

	// Readings
	rc := dataController.NewReadingController(dic)
	r.HandleFunc(common.ApiReadingCountRoute, authenticationHook(rc.ReadingTotalCount)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiAllReadingRoute, authenticationHook(rc.AllReadings)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiReadingByDeviceNameRoute, authenticationHook(rc.ReadingsByDeviceName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiReadingByTimeRangeRoute, authenticationHook(rc.ReadingsByTimeRange)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiReadingByResourceNameRoute, authenticationHook(rc.ReadingsByResourceName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiReadingCountByDeviceNameRoute, authenticationHook(rc.ReadingCountByDeviceName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiReadingByResourceNameAndTimeRangeRoute, authenticationHook(rc.ReadingsByResourceNameAndTimeRange)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiReadingByDeviceNameAndResourceNameRoute, authenticationHook(rc.ReadingsByDeviceNameAndResourceName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiReadingByDeviceNameAndResourceNameAndTimeRangeRoute, authenticationHook(rc.ReadingsByDeviceNameAndResourceNameAndTimeRange)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiReadingByDeviceNameAndTimeRangeRoute, authenticationHook(rc.ReadingsByDeviceNameAndResourceNamesAndTimeRange)).Methods(http.MethodGet)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.LoggingMiddleware(container.LoggingClientFrom(dic.Get)))
	r.Use(correlation.UrlDecodeMiddleware(container.LoggingClientFrom(dic.Get)))
}
