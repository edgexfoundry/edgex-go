//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
)

// DeviceServiceResponse defines the Response Content for GET DeviceService DTOs.
type DeviceServiceResponse struct {
	common.BaseResponse `json:",inline"`
	Service             dtos.DeviceService `json:"service"`
}

func NewDeviceServiceResponse(requestId string, message string, statusCode int, deviceService dtos.DeviceService) DeviceServiceResponse {
	return DeviceServiceResponse{
		BaseResponse: common.NewBaseResponse(requestId, message, statusCode),
		Service:      deviceService,
	}
}

// MultiDeviceServicesResponse defines the Response Content for GET multiple DeviceService DTOs.
type MultiDeviceServicesResponse struct {
	common.BaseWithTotalCountResponse `json:",inline"`
	Services                          []dtos.DeviceService `json:"services"`
}

func NewMultiDeviceServicesResponse(requestId string, message string, statusCode int, totalCount uint32, deviceServices []dtos.DeviceService) MultiDeviceServicesResponse {
	return MultiDeviceServicesResponse{
		BaseWithTotalCountResponse: common.NewBaseWithTotalCountResponse(requestId, message, statusCode, totalCount),
		Services:                   deviceServices,
	}
}
