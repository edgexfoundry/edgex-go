// Copyright (C) 2020-2025 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package notifications

import (
	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/controller"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"

	notificationsController "github.com/edgexfoundry/edgex-go/internal/support/notifications/controller/http"

	"github.com/labstack/echo/v4"
)

func LoadRestRoutes(r *echo.Echo, dic *di.Container, serviceName string) {
	authenticationHook := handlers.AutoConfigAuthenticationFunc(dic)

	// Common
	_ = controller.NewCommonController(dic, r, serviceName, edgex.Version)

	// Subscription
	sc := notificationsController.NewSubscriptionController(dic)
	r.POST(common.ApiSubscriptionRoute, sc.AddSubscription, authenticationHook)
	r.GET(common.ApiAllSubscriptionRoute, sc.AllSubscriptions, authenticationHook)
	r.GET(common.ApiSubscriptionByNameRoute, sc.SubscriptionByName, authenticationHook)
	r.GET(common.ApiSubscriptionByCategoryRoute, sc.SubscriptionsByCategory, authenticationHook)
	r.GET(common.ApiSubscriptionByLabelRoute, sc.SubscriptionsByLabel, authenticationHook)
	r.GET(common.ApiSubscriptionByReceiverRoute, sc.SubscriptionsByReceiver, authenticationHook)
	r.DELETE(common.ApiSubscriptionByNameRoute, sc.DeleteSubscriptionByName, authenticationHook)
	r.PATCH(common.ApiSubscriptionRoute, sc.PatchSubscription, authenticationHook)

	// Notification
	nc := notificationsController.NewNotificationController(dic)
	r.POST(common.ApiNotificationRoute, nc.AddNotification, authenticationHook)
	r.GET(common.ApiNotificationRoute, nc.NotificationsByQueryConditions, authenticationHook)
	r.GET(common.ApiNotificationByIdRoute, nc.NotificationById, authenticationHook)
	r.DELETE(common.ApiNotificationByIdRoute, nc.DeleteNotificationById, authenticationHook)
	r.DELETE(common.ApiNotificationByIdsRoute, nc.DeleteNotificationByIds, authenticationHook)
	r.GET(common.ApiNotificationByCategoryRoute, nc.NotificationsByCategory, authenticationHook)
	r.GET(common.ApiNotificationByLabelRoute, nc.NotificationsByLabel, authenticationHook)
	r.GET(common.ApiNotificationByStatusRoute, nc.NotificationsByStatus, authenticationHook)
	r.GET(common.ApiNotificationByTimeRangeRoute, nc.NotificationsByTimeRange, authenticationHook)
	r.GET(common.ApiNotificationBySubscriptionNameRoute, nc.NotificationsBySubscriptionName, authenticationHook)
	r.DELETE(common.ApiNotificationCleanupByAgeRoute, nc.CleanupNotificationsByAge, authenticationHook)
	r.DELETE(common.ApiNotificationCleanupRoute, nc.CleanupNotifications, authenticationHook)
	r.DELETE(common.ApiNotificationByAgeRoute, nc.DeleteProcessedNotificationsByAge, authenticationHook)
	r.PUT(common.ApiNotificationAcknowledgeByIdsRoute, nc.AcknowledgeNotificationByIds, authenticationHook)
	r.PUT(common.ApiNotificationUnacknowledgeByIdsRoute, nc.UnacknowledgeNotificationByIds, authenticationHook)

	// Transmission
	trans := notificationsController.NewTransmissionController(dic)
	r.GET(common.ApiTransmissionByIdRoute, trans.TransmissionById, authenticationHook)
	r.GET(common.ApiTransmissionByTimeRangeRoute, trans.TransmissionsByTimeRange, authenticationHook)
	r.GET(common.ApiAllTransmissionRoute, trans.AllTransmissions, authenticationHook)
	r.GET(common.ApiTransmissionByStatusRoute, trans.TransmissionsByStatus, authenticationHook)
	r.DELETE(common.ApiTransmissionByAgeRoute, trans.DeleteProcessedTransmissionsByAge, authenticationHook)
	r.GET(common.ApiTransmissionBySubscriptionNameRoute, trans.TransmissionsBySubscriptionName, authenticationHook)
	r.GET(common.ApiTransmissionByNotificationIdRoute, trans.TransmissionsByNotificationId, authenticationHook)
}
