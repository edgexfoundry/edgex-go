//
// Copyright (C) 2020-2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
)

// DeviceProfileResponse defines the Response Content for GET DeviceProfile DTOs.
type DeviceProfileResponse struct {
	common.BaseResponse `json:",inline"`
	Profile             dtos.DeviceProfile `json:"profile"`
}

func NewDeviceProfileResponse(requestId string, message string, statusCode int, deviceProfile dtos.DeviceProfile) DeviceProfileResponse {
	return DeviceProfileResponse{
		BaseResponse: common.NewBaseResponse(requestId, message, statusCode),
		Profile:      deviceProfile,
	}
}

// MultiDeviceProfilesResponse defines the Response Content for GET multiple DeviceProfile DTOs.
type MultiDeviceProfilesResponse struct {
	common.BaseWithTotalCountResponse `json:",inline"`
	Profiles                          []dtos.DeviceProfile `json:"profiles"`
}

func NewMultiDeviceProfilesResponse(requestId string, message string, statusCode int, totalCount uint32, deviceProfiles []dtos.DeviceProfile) MultiDeviceProfilesResponse {
	return MultiDeviceProfilesResponse{
		BaseWithTotalCountResponse: common.NewBaseWithTotalCountResponse(requestId, message, statusCode, totalCount),
		Profiles:                   deviceProfiles,
	}
}

// MultiDeviceProfileBasicInfoResponse defines the Response Content for GET multiple DeviceProfileBasicInfo DTOs.
type MultiDeviceProfileBasicInfoResponse struct {
	common.BaseWithTotalCountResponse `json:",inline"`
	Profiles                          []dtos.DeviceProfileBasicInfo `json:"profiles"`
}

func NewMultiDeviceProfileBasicInfosResponse(requestId string, message string, statusCode int, totalCount uint32, deviceProfileBasicInfos []dtos.DeviceProfileBasicInfo) MultiDeviceProfileBasicInfoResponse {
	return MultiDeviceProfileBasicInfoResponse{
		BaseWithTotalCountResponse: common.NewBaseWithTotalCountResponse(requestId, message, statusCode, totalCount),
		Profiles:                   deviceProfileBasicInfos,
	}
}
