// Copyright (C) 2020-2021 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package notifications

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/gorilla/mux"

	commonController "github.com/edgexfoundry/edgex-go/internal/pkg/controller/http"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	notificationsController "github.com/edgexfoundry/edgex-go/internal/support/notifications/controller/http"
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
	r.HandleFunc(common.ApiSecretRoute, authenticationHook(cc.AddSecret)).Methods(http.MethodPost)

	// Subscription
	sc := notificationsController.NewSubscriptionController(dic)
	r.HandleFunc(common.ApiSubscriptionRoute, authenticationHook(sc.AddSubscription)).Methods(http.MethodPost)
	r.HandleFunc(common.ApiAllSubscriptionRoute, authenticationHook(sc.AllSubscriptions)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiSubscriptionByNameRoute, authenticationHook(sc.SubscriptionByName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiSubscriptionByCategoryRoute, authenticationHook(sc.SubscriptionsByCategory)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiSubscriptionByLabelRoute, authenticationHook(sc.SubscriptionsByLabel)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiSubscriptionByReceiverRoute, authenticationHook(sc.SubscriptionsByReceiver)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiSubscriptionByNameRoute, authenticationHook(sc.DeleteSubscriptionByName)).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiSubscriptionRoute, authenticationHook(sc.PatchSubscription)).Methods(http.MethodPatch)

	// Notification
	nc := notificationsController.NewNotificationController(dic)
	r.HandleFunc(common.ApiNotificationRoute, authenticationHook(nc.AddNotification)).Methods(http.MethodPost)
	r.HandleFunc(common.ApiNotificationByIdRoute, authenticationHook(nc.NotificationById)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiNotificationByIdRoute, authenticationHook(nc.DeleteNotificationById)).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiNotificationByCategoryRoute, authenticationHook(nc.NotificationsByCategory)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiNotificationByLabelRoute, authenticationHook(nc.NotificationsByLabel)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiNotificationByStatusRoute, authenticationHook(nc.NotificationsByStatus)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiNotificationByTimeRangeRoute, authenticationHook(nc.NotificationsByTimeRange)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiNotificationBySubscriptionNameRoute, authenticationHook(nc.NotificationsBySubscriptionName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiNotificationCleanupByAgeRoute, authenticationHook(nc.CleanupNotificationsByAge)).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiNotificationCleanupRoute, authenticationHook(nc.CleanupNotifications)).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiNotificationByAgeRoute, authenticationHook(nc.DeleteProcessedNotificationsByAge)).Methods(http.MethodDelete)

	// Transmission
	trans := notificationsController.NewTransmissionController(dic)
	r.HandleFunc(common.ApiTransmissionByIdRoute, authenticationHook(trans.TransmissionById)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiTransmissionByTimeRangeRoute, authenticationHook(trans.TransmissionsByTimeRange)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiAllTransmissionRoute, authenticationHook(trans.AllTransmissions)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiTransmissionByStatusRoute, authenticationHook(trans.TransmissionsByStatus)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiTransmissionByAgeRoute, authenticationHook(trans.DeleteProcessedTransmissionsByAge)).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiTransmissionBySubscriptionNameRoute, authenticationHook(trans.TransmissionsBySubscriptionName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiTransmissionByNotificationIdRoute, authenticationHook(trans.TransmissionsByNotificationId)).Methods(http.MethodGet)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.LoggingMiddleware(container.LoggingClientFrom(dic.Get)))
}
