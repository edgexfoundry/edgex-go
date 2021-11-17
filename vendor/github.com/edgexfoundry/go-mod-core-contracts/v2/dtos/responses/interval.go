//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
)

// IntervalResponse defines the Response Content for GET Interval DTOs.
// This object and its properties correspond to the IntervalResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/support-scheduler/2.1.0#/IntervalResponse
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
// This object and its properties correspond to the MultiIntervalsResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/support-scheduler/2.1.0#/MultiIntervalsResponse
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
