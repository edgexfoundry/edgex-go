//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
)

// IntervalActionResponse defines the Response Content for GET IntervalAction DTOs.
type IntervalActionResponse struct {
	common.BaseResponse `json:",inline"`
	Action              dtos.IntervalAction `json:"action"`
}

func NewIntervalActionResponse(requestId string, message string, statusCode int, action dtos.IntervalAction) IntervalActionResponse {
	return IntervalActionResponse{
		BaseResponse: common.NewBaseResponse(requestId, message, statusCode),
		Action:       action,
	}
}

// MultiIntervalActionsResponse defines the Response Content for GET multiple IntervalAction DTOs.
type MultiIntervalActionsResponse struct {
	common.BaseWithTotalCountResponse `json:",inline"`
	Actions                           []dtos.IntervalAction `json:"actions"`
}

func NewMultiIntervalActionsResponse(requestId string, message string, statusCode int, totalCount uint32, actions []dtos.IntervalAction) MultiIntervalActionsResponse {
	return MultiIntervalActionsResponse{
		BaseWithTotalCountResponse: common.NewBaseWithTotalCountResponse(requestId, message, statusCode, totalCount),
		Actions:                    actions,
	}
}
