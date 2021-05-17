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

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	commonController "github.com/edgexfoundry/edgex-go/internal/pkg/v2/controller/http"
	schedulerController "github.com/edgexfoundry/edgex-go/internal/support/scheduler/v2/controller/http"
)

func LoadRestRoutes(r *mux.Router, dic *di.Container) {
	// v2 API routes
	// Common
	cc := commonController.NewV2CommonController(dic)
	r.HandleFunc(v2.ApiPingRoute, cc.Ping).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiVersionRoute, cc.Version).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiConfigRoute, cc.Config).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiMetricsRoute, cc.Metrics).Methods(http.MethodGet)

	// Interval
	interval := schedulerController.NewIntervalController(dic)
	r.HandleFunc(v2.ApiIntervalRoute, interval.AddInterval).Methods(http.MethodPost)
	r.HandleFunc(v2.ApiIntervalByNameRoute, interval.IntervalByName).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiAllIntervalRoute, interval.AllIntervals).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiIntervalByNameRoute, interval.DeleteIntervalByName).Methods(http.MethodDelete)
	r.HandleFunc(v2.ApiIntervalRoute, interval.PatchInterval).Methods(http.MethodPatch)

	// IntervalAction
	action := schedulerController.NewIntervalActionController(dic)
	r.HandleFunc(v2.ApiIntervalActionRoute, action.AddIntervalAction).Methods(http.MethodPost)
	r.HandleFunc(v2.ApiAllIntervalActionRoute, action.AllIntervalActions).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiIntervalActionByNameRoute, action.IntervalActionByName).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiIntervalActionByNameRoute, action.DeleteIntervalActionByName).Methods(http.MethodDelete)
	r.HandleFunc(v2.ApiIntervalActionRoute, action.PatchIntervalAction).Methods(http.MethodPatch)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.LoggingMiddleware(container.LoggingClientFrom(dic.Get)))
}
