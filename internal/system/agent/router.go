//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"

	commonController "github.com/edgexfoundry/edgex-go/internal/pkg/controller/http"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	smaController "github.com/edgexfoundry/edgex-go/internal/system/agent/controller/http"
)

func LoadRestRoutes(r *mux.Router, dic *di.Container, serviceName string) {
	// Common
	cc := commonController.NewCommonController(dic, serviceName)
	r.HandleFunc(common.ApiPingRoute, cc.Ping).Methods(http.MethodGet)
	r.HandleFunc(common.ApiVersionRoute, cc.Version).Methods(http.MethodGet)
	r.HandleFunc(common.ApiConfigRoute, cc.Config).Methods(http.MethodGet)
	r.HandleFunc(common.ApiMetricsRoute, cc.Metrics).Methods(http.MethodGet)

	ac := smaController.NewAgentController(dic)
	r.HandleFunc(common.ApiHealthRoute, ac.GetHealth).Methods(http.MethodGet)
	r.HandleFunc(common.ApiMultiMetricsRoute, ac.GetMetrics).Methods(http.MethodGet)
	r.HandleFunc(common.ApiMultiConfigRoute, ac.GetConfigs).Methods(http.MethodGet)
	r.HandleFunc(common.ApiOperationRoute, ac.PostOperations).Methods(http.MethodPost)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.LoggingMiddleware(container.LoggingClientFrom(dic.Get)))
}
