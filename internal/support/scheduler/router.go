// Copyright (C) 2021 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/gorilla/mux"

	commonController "github.com/edgexfoundry/edgex-go/internal/pkg/controller/http"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	schedulerController "github.com/edgexfoundry/edgex-go/internal/support/scheduler/controller/http"
)

func LoadRestRoutes(r *mux.Router, dic *di.Container, serviceName string) {
	lc := container.LoggingClientFrom(dic.Get)
	secretProvider := container.SecretProviderExtFrom(dic.Get)
	authenticationHook := handlers.AutoConfigAuthenticationFunc(secretProvider, lc)

	// Common
	cc := commonController.NewCommonController(dic, serviceName)
	r.HandleFunc(common.ApiPingRoute, cc.Ping).Methods(http.MethodGet) // Health check is always unauthenticated
	r.HandleFunc(common.ApiVersionRoute, authenticationHook(cc.Version)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiConfigRoute, authenticationHook(cc.Config)).Methods(http.MethodGet)

	// Interval
	interval := schedulerController.NewIntervalController(dic)
	r.HandleFunc(common.ApiIntervalRoute, authenticationHook(interval.AddInterval)).Methods(http.MethodPost)
	r.HandleFunc(common.ApiIntervalByNameRoute, authenticationHook(interval.IntervalByName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiAllIntervalRoute, authenticationHook(interval.AllIntervals)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiIntervalByNameRoute, authenticationHook(interval.DeleteIntervalByName)).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiIntervalRoute, authenticationHook(interval.PatchInterval)).Methods(http.MethodPatch)

	// IntervalAction
	action := schedulerController.NewIntervalActionController(dic)
	r.HandleFunc(common.ApiIntervalActionRoute, authenticationHook(action.AddIntervalAction)).Methods(http.MethodPost)
	r.HandleFunc(common.ApiAllIntervalActionRoute, authenticationHook(action.AllIntervalActions)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiIntervalActionByNameRoute, authenticationHook(action.IntervalActionByName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiIntervalActionByNameRoute, authenticationHook(action.DeleteIntervalActionByName)).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiIntervalActionRoute, authenticationHook(action.PatchIntervalAction)).Methods(http.MethodPatch)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.LoggingMiddleware(container.LoggingClientFrom(dic.Get)))
}
