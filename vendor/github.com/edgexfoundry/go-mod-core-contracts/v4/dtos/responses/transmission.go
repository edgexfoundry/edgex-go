//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
)

// TransmissionResponse defines the Response Content for GET Transmission DTO.
type TransmissionResponse struct {
	common.BaseResponse `json:",inline"`
	Transmission        dtos.Transmission `json:"transmission"`
}

func NewTransmissionResponse(requestId string, message string, statusCode int,
	transmission dtos.Transmission) TransmissionResponse {
	return TransmissionResponse{
		BaseResponse: common.NewBaseResponse(requestId, message, statusCode),
		Transmission: transmission,
	}
}

// MultiTransmissionsResponse defines the Response Content for GET multiple Transmission DTOs.
type MultiTransmissionsResponse struct {
	common.BaseWithTotalCountResponse `json:",inline"`
	Transmissions                     []dtos.Transmission `json:"transmissions"`
}

func NewMultiTransmissionsResponse(requestId string, message string, statusCode int, totalCount uint32, transmissions []dtos.Transmission) MultiTransmissionsResponse {
	return MultiTransmissionsResponse{
		BaseWithTotalCountResponse: common.NewBaseWithTotalCountResponse(requestId, message, statusCode, totalCount),
		Transmissions:              transmissions,
	}
}
