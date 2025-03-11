//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
)

// NotificationResponse defines the Response Content for GET Notification DTO.
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
