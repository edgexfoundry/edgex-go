// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package notifications

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/gorilla/mux"

	commonController "github.com/edgexfoundry/edgex-go/internal/pkg/controller/http"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	notificationsController "github.com/edgexfoundry/edgex-go/internal/support/notifications/controller/http"
)

func LoadRestRoutes(r *mux.Router, dic *di.Container, serviceName string) {
	// Common
	cc := commonController.NewCommonController(dic, serviceName)
	r.HandleFunc(common.ApiPingRoute, cc.Ping).Methods(http.MethodGet)
	r.HandleFunc(common.ApiVersionRoute, cc.Version).Methods(http.MethodGet)
	r.HandleFunc(common.ApiConfigRoute, cc.Config).Methods(http.MethodGet)
	r.HandleFunc(common.ApiMetricsRoute, cc.Metrics).Methods(http.MethodGet)
	r.HandleFunc(common.ApiSecretRoute, cc.AddSecret).Methods(http.MethodPost)

	// Subscription
	sc := notificationsController.NewSubscriptionController(dic)
	r.HandleFunc(common.ApiSubscriptionRoute, sc.AddSubscription).Methods(http.MethodPost)
	r.HandleFunc(common.ApiAllSubscriptionRoute, sc.AllSubscriptions).Methods(http.MethodGet)
	r.HandleFunc(common.ApiSubscriptionByNameRoute, sc.SubscriptionByName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiSubscriptionByCategoryRoute, sc.SubscriptionsByCategory).Methods(http.MethodGet)
	r.HandleFunc(common.ApiSubscriptionByLabelRoute, sc.SubscriptionsByLabel).Methods(http.MethodGet)
	r.HandleFunc(common.ApiSubscriptionByReceiverRoute, sc.SubscriptionsByReceiver).Methods(http.MethodGet)
	r.HandleFunc(common.ApiSubscriptionByNameRoute, sc.DeleteSubscriptionByName).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiSubscriptionRoute, sc.PatchSubscription).Methods(http.MethodPatch)

	// Notification
	nc := notificationsController.NewNotificationController(dic)
	r.HandleFunc(common.ApiNotificationRoute, nc.AddNotification).Methods(http.MethodPost)
	r.HandleFunc(common.ApiNotificationByIdRoute, nc.NotificationById).Methods(http.MethodGet)
	r.HandleFunc(common.ApiNotificationByIdRoute, nc.DeleteNotificationById).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiNotificationByCategoryRoute, nc.NotificationsByCategory).Methods(http.MethodGet)
	r.HandleFunc(common.ApiNotificationByLabelRoute, nc.NotificationsByLabel).Methods(http.MethodGet)
	r.HandleFunc(common.ApiNotificationByStatusRoute, nc.NotificationsByStatus).Methods(http.MethodGet)
	r.HandleFunc(common.ApiNotificationByTimeRangeRoute, nc.NotificationsByTimeRange).Methods(http.MethodGet)
	r.HandleFunc(common.ApiNotificationBySubscriptionNameRoute, nc.NotificationsBySubscriptionName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiNotificationCleanupByAgeRoute, nc.CleanupNotificationsByAge).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiNotificationCleanupRoute, nc.CleanupNotifications).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiNotificationByAgeRoute, nc.DeleteProcessedNotificationsByAge).Methods(http.MethodDelete)

	// Transmission
	trans := notificationsController.NewTransmissionController(dic)
	r.HandleFunc(common.ApiTransmissionByIdRoute, trans.TransmissionById).Methods(http.MethodGet)
	r.HandleFunc(common.ApiTransmissionByTimeRangeRoute, trans.TransmissionsByTimeRange).Methods(http.MethodGet)
	r.HandleFunc(common.ApiAllTransmissionRoute, trans.AllTransmissions).Methods(http.MethodGet)
	r.HandleFunc(common.ApiTransmissionByStatusRoute, trans.TransmissionsByStatus).Methods(http.MethodGet)
	r.HandleFunc(common.ApiTransmissionByAgeRoute, trans.DeleteProcessedTransmissionsByAge).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiTransmissionBySubscriptionNameRoute, trans.TransmissionsBySubscriptionName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiTransmissionByNotificationIdRoute, trans.TransmissionsByNotificationId).Methods(http.MethodGet)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.LoggingMiddleware(container.LoggingClientFrom(dic.Get)))
}
