// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/gorilla/mux"

	commonController "github.com/edgexfoundry/edgex-go/internal/pkg/controller/http"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	schedulerController "github.com/edgexfoundry/edgex-go/internal/support/scheduler/controller/http"
)

func LoadRestRoutes(r *mux.Router, dic *di.Container, serviceName string) {
	// Common
	cc := commonController.NewCommonController(dic, serviceName)
	r.HandleFunc(common.ApiPingRoute, cc.Ping).Methods(http.MethodGet)
	r.HandleFunc(common.ApiVersionRoute, cc.Version).Methods(http.MethodGet)
	r.HandleFunc(common.ApiConfigRoute, cc.Config).Methods(http.MethodGet)
	r.HandleFunc(common.ApiMetricsRoute, cc.Metrics).Methods(http.MethodGet)

	// Interval
	interval := schedulerController.NewIntervalController(dic)
	r.HandleFunc(common.ApiIntervalRoute, interval.AddInterval).Methods(http.MethodPost)
	r.HandleFunc(common.ApiIntervalByNameRoute, interval.IntervalByName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiAllIntervalRoute, interval.AllIntervals).Methods(http.MethodGet)
	r.HandleFunc(common.ApiIntervalByNameRoute, interval.DeleteIntervalByName).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiIntervalRoute, interval.PatchInterval).Methods(http.MethodPatch)

	// IntervalAction
	action := schedulerController.NewIntervalActionController(dic)
	r.HandleFunc(common.ApiIntervalActionRoute, action.AddIntervalAction).Methods(http.MethodPost)
	r.HandleFunc(common.ApiAllIntervalActionRoute, action.AllIntervalActions).Methods(http.MethodGet)
	r.HandleFunc(common.ApiIntervalActionByNameRoute, action.IntervalActionByName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiIntervalActionByNameRoute, action.DeleteIntervalActionByName).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiIntervalActionRoute, action.PatchIntervalAction).Methods(http.MethodPatch)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.LoggingMiddleware(container.LoggingClientFrom(dic.Get)))
}
