//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
)

// IntervalResponse defines the Response Content for GET Interval DTOs.
type IntervalResponse struct {
	common.BaseResponse `json:",inline"`
	Interval            dtos.Interval `json:"interval"`
}

func NewIntervalResponse(requestId string, message string, statusCode int, interval dtos.Interval) IntervalResponse {
	return IntervalResponse{
		BaseResponse: common.NewBaseResponse(requestId, message, statusCode),
		Interval:     interval,
	}
}

// MultiIntervalsResponse defines the Response Content for GET multiple Interval DTOs.
type MultiIntervalsResponse struct {
	common.BaseWithTotalCountResponse `json:",inline"`
	Intervals                         []dtos.Interval `json:"intervals"`
}

func NewMultiIntervalsResponse(requestId string, message string, statusCode int, totalCount uint32, intervals []dtos.Interval) MultiIntervalsResponse {
	return MultiIntervalsResponse{
		BaseWithTotalCountResponse: common.NewBaseWithTotalCountResponse(requestId, message, statusCode, totalCount),
		Intervals:                  intervals,
	}
}
