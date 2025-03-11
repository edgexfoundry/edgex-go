//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
)

// SubscriptionResponse defines the Subscription Content for GET Subscription DTOs.
type SubscriptionResponse struct {
	common.BaseResponse `json:",inline"`
	Subscription        dtos.Subscription `json:"subscription"`
}

func NewSubscriptionResponse(requestId string, message string, statusCode int,
	subscription dtos.Subscription) SubscriptionResponse {
	return SubscriptionResponse{
		BaseResponse: common.NewBaseResponse(requestId, message, statusCode),
		Subscription: subscription,
	}
}

// MultiSubscriptionsResponse defines the Subscription Content for GET multiple Subscription DTOs.
type MultiSubscriptionsResponse struct {
	common.BaseWithTotalCountResponse `json:",inline"`
	Subscriptions                     []dtos.Subscription `json:"subscriptions"`
}

func NewMultiSubscriptionsResponse(requestId string, message string, statusCode int, totalCount uint32, subscriptions []dtos.Subscription) MultiSubscriptionsResponse {
	return MultiSubscriptionsResponse{
		BaseWithTotalCountResponse: common.NewBaseWithTotalCountResponse(requestId, message, statusCode, totalCount),
		Subscriptions:              subscriptions,
	}
}
