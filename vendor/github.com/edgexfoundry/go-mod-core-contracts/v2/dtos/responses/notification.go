//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
)

// NotificationResponse defines the Response Content for GET Notification DTO.
// This object and its properties correspond to the NotificationResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/support-notifications/2.1.0#/NotificationResponse
type NotificationResponse struct {
	common.BaseResponse `json:",inline"`
	Notification        dtos.Notification `json:"notification"`
}

func NewNotificationResponse(requestId string, message string, statusCode int,
	notification dtos.Notification) NotificationResponse {
	return NotificationResponse{
		BaseResponse: common.NewBaseResponse(requestId, message, statusCode),
		Notification: notification,
	}
}

// MultiNotificationsResponse defines the Response Content for GET multiple Notification DTOs.
// This object and its properties correspond to the MultiNotificationsResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/support-notifications/2.1.0#/MultiNotificationsResponse
type MultiNotificationsResponse struct {
	common.BaseWithTotalCountResponse `json:",inline"`
	Notifications                     []dtos.Notification `json:"notifications"`
}

func NewMultiNotificationsResponse(requestId string, message string, statusCode int, totalCount uint32, notifications []dtos.Notification) MultiNotificationsResponse {
	return MultiNotificationsResponse{
		BaseWithTotalCountResponse: common.NewBaseWithTotalCountResponse(requestId, message, statusCode, totalCount),
		Notifications:              notifications,
	}
}
