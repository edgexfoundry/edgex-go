// Copyright (C) 2020-2021 IOTech Ltd
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
	notificationsController "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/controller/http"
)

func LoadRestRoutes(r *mux.Router, dic *di.Container) {
	// v2 API routes
	// Common
	cc := commonController.NewV2CommonController(dic)
	r.HandleFunc(v2.ApiPingRoute, cc.Ping).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiVersionRoute, cc.Version).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiConfigRoute, cc.Config).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiMetricsRoute, cc.Metrics).Methods(http.MethodGet)

	// Subscription
	sc := notificationsController.NewSubscriptionController(dic)
	r.HandleFunc(v2.ApiSubscriptionRoute, sc.AddSubscription).Methods(http.MethodPost)
	r.HandleFunc(v2.ApiAllSubscriptionRoute, sc.AllSubscriptions).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiSubscriptionByNameRoute, sc.SubscriptionByName).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiSubscriptionByCategoryRoute, sc.SubscriptionsByCategory).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiSubscriptionByLabelRoute, sc.SubscriptionsByLabel).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiSubscriptionByReceiverRoute, sc.SubscriptionsByReceiver).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiSubscriptionByNameRoute, sc.DeleteSubscriptionByName).Methods(http.MethodDelete)
	r.HandleFunc(v2.ApiSubscriptionRoute, sc.PatchSubscription).Methods(http.MethodPatch)

	// Notification
	nc := notificationsController.NewNotificationController(dic)
	r.HandleFunc(v2.ApiNotificationRoute, nc.AddNotification).Methods(http.MethodPost)
	r.HandleFunc(v2.ApiNotificationByIdRoute, nc.NotificationById).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiNotificationByIdRoute, nc.DeleteNotificationById).Methods(http.MethodDelete)
	r.HandleFunc(v2.ApiNotificationByCategoryRoute, nc.NotificationsByCategory).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiNotificationByLabelRoute, nc.NotificationsByLabel).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiNotificationByStatusRoute, nc.NotificationsByStatus).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiNotificationByTimeRangeRoute, nc.NotificationsByTimeRange).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiNotificationBySubscriptionNameRoute, nc.NotificationsBySubscriptionName).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiNotificationCleanupByAgeRoute, nc.CleanupNotificationsByAge).Methods(http.MethodDelete)
	r.HandleFunc(v2.ApiNotificationCleanupRoute, nc.CleanupNotifications).Methods(http.MethodDelete)
	r.HandleFunc(v2.ApiNotificationByAgeRoute, nc.DeleteProcessedNotificationsByAge).Methods(http.MethodDelete)

	// Transmission
	trans := notificationsController.NewTransmissionController(dic)
	r.HandleFunc(v2.ApiTransmissionByIdRoute, trans.TransmissionById).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiTransmissionByTimeRangeRoute, trans.TransmissionsByTimeRange).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiAllTransmissionRoute, trans.AllTransmissions).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiTransmissionByStatusRoute, trans.TransmissionsByStatus).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiTransmissionByAgeRoute, trans.DeleteProcessedTransmissionsByAge).Methods(http.MethodDelete)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.LoggingMiddleware(container.LoggingClientFrom(dic.Get)))
}
