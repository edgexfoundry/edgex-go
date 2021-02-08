// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package v2

import (
	"net/http"

	commonController "github.com/edgexfoundry/edgex-go/internal/pkg/v2/controller/http"
	notificationsController "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/controller/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	v2Constant "github.com/edgexfoundry/go-mod-core-contracts/v2/v2"

	"github.com/gorilla/mux"
)

func LoadRestRoutes(r *mux.Router, dic *di.Container) {
	// v2 API routes
	// Common
	cc := commonController.NewV2CommonController(dic)
	r.HandleFunc(v2Constant.ApiPingRoute, cc.Ping).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiVersionRoute, cc.Version).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiConfigRoute, cc.Config).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiMetricsRoute, cc.Metrics).Methods(http.MethodGet)

	// Subscription
	nc := notificationsController.NewSubscriptionController(dic)
	r.HandleFunc(v2Constant.ApiSubscriptionRoute, nc.AddSubscription).Methods(http.MethodPost)
	r.HandleFunc(v2Constant.ApiAllSubscriptionRoute, nc.AllSubscriptions).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiSubscriptionByNameRoute, nc.SubscriptionByName).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiSubscriptionByCategoryRoute, nc.SubscriptionsByCategory).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiSubscriptionByLabelRoute, nc.SubscriptionsByLabel).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiSubscriptionByReceiverRoute, nc.SubscriptionsByReceiver).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiSubscriptionByNameRoute, nc.DeleteSubscriptionByName).Methods(http.MethodDelete)
	r.HandleFunc(v2Constant.ApiSubscriptionRoute, nc.PatchSubscription).Methods(http.MethodPatch)
}
