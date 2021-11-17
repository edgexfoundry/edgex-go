//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
)

// DeviceResourceResponse defines the Response Content for GET DeviceResource DTOs.
// This object and its properties correspond to the DeviceResourceResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-metadata/2.1.0#/DeviceResourceResponse
type DeviceResourceResponse struct {
	common.BaseResponse `json:",inline"`
	Resource            dtos.DeviceResource `json:"resource"`
}

// NewDeviceResourceResponse creates deviceResource response DTO with required fields
func NewDeviceResourceResponse(requestId string, message string, statusCode int, resource dtos.DeviceResource) DeviceResourceResponse {
	return DeviceResourceResponse{
		BaseResponse: common.NewBaseResponse(requestId, message, statusCode),
		Resource:     resource,
	}
}
