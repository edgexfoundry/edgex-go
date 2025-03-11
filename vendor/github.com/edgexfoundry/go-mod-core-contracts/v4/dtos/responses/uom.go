//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"

type UnitsOfMeasureResponse struct {
	common.BaseResponse `json:",inline"`
	Uom                 any `json:"uom"`
}

func NewUnitsOfMeasureResponse(requestId string, message string, statusCode int, uom any) UnitsOfMeasureResponse {
	return UnitsOfMeasureResponse{
		BaseResponse: common.NewBaseResponse(requestId, message, statusCode),
		Uom:          uom,
	}
}
