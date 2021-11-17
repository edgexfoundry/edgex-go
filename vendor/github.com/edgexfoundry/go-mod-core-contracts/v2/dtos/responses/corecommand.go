//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
)

// DeviceCoreCommandResponse defines the Response Content for GET DeviceCoreCommand DTO.
// This object and its properties correspond to the DeviceCoreCommandResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-command/2.1.0#/DeviceCoreCommandResponse
type DeviceCoreCommandResponse struct {
	common.BaseResponse `json:",inline"`
	DeviceCoreCommand   dtos.DeviceCoreCommand `json:"deviceCoreCommand"`
}

func NewDeviceCoreCommandResponse(requestId string, message string, statusCode int, deviceCoreCommand dtos.DeviceCoreCommand) DeviceCoreCommandResponse {
	return DeviceCoreCommandResponse{
		BaseResponse:      common.NewBaseResponse(requestId, message, statusCode),
		DeviceCoreCommand: deviceCoreCommand,
	}
}

// MultiDeviceCoreCommandsResponse defines the Response Content for GET multiple DeviceCoreCommand DTOs.
// This object and its properties correspond to the MultiDeviceCoreCommandsResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-command/2.1.0#/MultiDeviceCoreCommandsResponse
type MultiDeviceCoreCommandsResponse struct {
	common.BaseWithTotalCountResponse `json:",inline"`
	DeviceCoreCommands                []dtos.DeviceCoreCommand `json:"deviceCoreCommands"`
}

func NewMultiDeviceCoreCommandsResponse(requestId string, message string, statusCode int, totalCount uint32, commands []dtos.DeviceCoreCommand) MultiDeviceCoreCommandsResponse {
	return MultiDeviceCoreCommandsResponse{
		BaseWithTotalCountResponse: common.NewBaseWithTotalCountResponse(requestId, message, statusCode, totalCount),
		DeviceCoreCommands:         commands,
	}
}
