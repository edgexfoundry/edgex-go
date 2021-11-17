//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
)

// ReadingResponse defines the Response Content for GET reading DTO.
// This object and its properties correspond to the ReadingResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-data/2.1.0#/ReadingResponse
type ReadingResponse struct {
	common.BaseResponse `json:",inline"`
	Reading             dtos.BaseReading `json:"reading"`
}

// MultiReadingsResponse defines the Response Content for GET multiple reading DTO.
// This object and its properties correspond to the MultiReadingsResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-data/2.1.0#/MultiReadingsResponse
type MultiReadingsResponse struct {
	common.BaseWithTotalCountResponse `json:",inline"`
	Readings                          []dtos.BaseReading `json:"readings"`
}

func NewReadingResponse(requestId string, message string, statusCode int, reading dtos.BaseReading) ReadingResponse {
	return ReadingResponse{
		BaseResponse: common.NewBaseResponse(requestId, message, statusCode),
		Reading:      reading,
	}
}

func NewMultiReadingsResponse(requestId string, message string, statusCode int, totalCount uint32, readings []dtos.BaseReading) MultiReadingsResponse {
	return MultiReadingsResponse{
		BaseWithTotalCountResponse: common.NewBaseWithTotalCountResponse(requestId, message, statusCode, totalCount),
		Readings:                   readings,
	}
}
