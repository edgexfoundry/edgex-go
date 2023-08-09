// Copyright (C) 2020-2023 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package notifications

import (
	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/controller"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"

	notificationsController "github.com/edgexfoundry/edgex-go/internal/support/notifications/controller/http"

	"github.com/labstack/echo/v4"
)

func LoadRestRoutes(r *echo.Echo, dic *di.Container, serviceName string) {
	lc := container.LoggingClientFrom(dic.Get)
	secretProvider := container.SecretProviderExtFrom(dic.Get)
	authenticationHook := handlers.AutoConfigAuthenticationFunc(secretProvider, lc)

	// Common
	_ = controller.NewCommonController(dic, r, serviceName, edgex.Version)

	// Subscription
	sc := notificationsController.NewSubscriptionController(dic)
	r.POST(common.ApiSubscriptionRoute, sc.AddSubscription, authenticationHook)
	r.GET(common.ApiAllSubscriptionRoute, sc.AllSubscriptions, authenticationHook)
	r.GET(common.ApiSubscriptionByNameEchoRoute, sc.SubscriptionByName, authenticationHook)
	r.GET(common.ApiSubscriptionByCategoryEchoRoute, sc.SubscriptionsByCategory, authenticationHook)
	r.GET(common.ApiSubscriptionByLabelEchoRoute, sc.SubscriptionsByLabel, authenticationHook)
	r.GET(common.ApiSubscriptionByReceiverEchoRoute, sc.SubscriptionsByReceiver, authenticationHook)
	r.DELETE(common.ApiSubscriptionByNameEchoRoute, sc.DeleteSubscriptionByName, authenticationHook)
	r.PATCH(common.ApiSubscriptionRoute, sc.PatchSubscription, authenticationHook)

	// Notification
	nc := notificationsController.NewNotificationController(dic)
	r.POST(common.ApiNotificationRoute, nc.AddNotification, authenticationHook)
	r.GET(common.ApiNotificationByIdEchoRoute, nc.NotificationById, authenticationHook)
	r.DELETE(common.ApiNotificationByIdEchoRoute, nc.DeleteNotificationById, authenticationHook)
	r.GET(common.ApiNotificationByCategoryEchoRoute, nc.NotificationsByCategory, authenticationHook)
	r.GET(common.ApiNotificationByLabelEchoRoute, nc.NotificationsByLabel, authenticationHook)
	r.GET(common.ApiNotificationByStatusEchoRoute, nc.NotificationsByStatus, authenticationHook)
	r.GET(common.ApiNotificationByTimeRangeEchoRoute, nc.NotificationsByTimeRange, authenticationHook)
	r.GET(common.ApiNotificationBySubscriptionNameEchoRoute, nc.NotificationsBySubscriptionName, authenticationHook)
	r.DELETE(common.ApiNotificationCleanupByAgeEchoRoute, nc.CleanupNotificationsByAge, authenticationHook)
	r.DELETE(common.ApiNotificationCleanupRoute, nc.CleanupNotifications, authenticationHook)
	r.DELETE(common.ApiNotificationByAgeEchoRoute, nc.DeleteProcessedNotificationsByAge, authenticationHook)

	// Transmission
	trans := notificationsController.NewTransmissionController(dic)
	r.GET(common.ApiTransmissionByIdEchoRoute, trans.TransmissionById, authenticationHook)
	r.GET(common.ApiTransmissionByTimeRangeEchoRoute, trans.TransmissionsByTimeRange, authenticationHook)
	r.GET(common.ApiAllTransmissionRoute, trans.AllTransmissions, authenticationHook)
	r.GET(common.ApiTransmissionByStatusEchoRoute, trans.TransmissionsByStatus, authenticationHook)
	r.DELETE(common.ApiTransmissionByAgeEchoRoute, trans.DeleteProcessedTransmissionsByAge, authenticationHook)
	r.GET(common.ApiTransmissionBySubscriptionNameEchoRoute, trans.TransmissionsBySubscriptionName, authenticationHook)
	r.GET(common.ApiTransmissionByNotificationIdEchoRoute, trans.TransmissionsByNotificationId, authenticationHook)
}
