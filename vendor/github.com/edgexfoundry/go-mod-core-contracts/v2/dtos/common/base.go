//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/google/uuid"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
)

// BaseRequest defines the base content for request DTOs (data transfer objects).
// This object and its properties correspond to the BaseRequest object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-data/2.1.0#/BaseRequest
type BaseRequest struct {
	Versionable `json:",inline"`
	RequestId   string `json:"requestId" validate:"len=0|uuid"`
}

func NewBaseRequest() BaseRequest {
	return BaseRequest{
		Versionable: NewVersionable(),
		RequestId:   uuid.NewString(),
	}
}

// BaseResponse defines the base content for response DTOs (data transfer objects).
// This object and its properties correspond to the BaseResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-data/2.1.0#/BaseResponse
type BaseResponse struct {
	Versionable `json:",inline"`
	RequestId   string `json:"requestId,omitempty"`
	Message     string `json:"message,omitempty"`
	StatusCode  int    `json:"statusCode"`
}

// Versionable shows the API version in DTOs
type Versionable struct {
	ApiVersion string `json:"apiVersion" yaml:"apiVersion" validate:"required"`
}

// BaseWithIdResponse defines the base content for response DTOs (data transfer objects).
// This object and its properties correspond to the BaseWithIdResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-data/2.1.0#/BaseWithIdResponse
type BaseWithIdResponse struct {
	BaseResponse `json:",inline"`
	Id           string `json:"id"`
}

func NewBaseResponse(requestId string, message string, statusCode int) BaseResponse {
	return BaseResponse{
		Versionable: NewVersionable(),
		RequestId:   requestId,
		Message:     message,
		StatusCode:  statusCode,
	}
}

func NewVersionable() Versionable {
	return Versionable{ApiVersion: common.ApiVersion}
}

func NewBaseWithIdResponse(requestId string, message string, statusCode int, id string) BaseWithIdResponse {
	return BaseWithIdResponse{
		BaseResponse: NewBaseResponse(requestId, message, statusCode),
		Id:           id,
	}
}

// BaseWithTotalCountResponse defines the base content for response DTOs (data transfer objects).
// This object and its properties correspond to the BaseWithTotalCountResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-data/2.1.0#/BaseWithTotalCountResponse
type BaseWithTotalCountResponse struct {
	BaseResponse `json:",inline"`
	TotalCount   uint32 `json:"totalCount"`
}

func NewBaseWithTotalCountResponse(requestId string, message string, statusCode int, totalCount uint32) BaseWithTotalCountResponse {
	return BaseWithTotalCountResponse{
		BaseResponse: NewBaseResponse(requestId, message, statusCode),
		TotalCount:   totalCount,
	}
}

// BaseWithServiceNameResponse defines the base content for response DTOs (data transfer objects).
// This object and its properties correspond to the BaseWithServiceNameResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/system-agent/2.1.0#/BaseWithServiceNameResponse
type BaseWithServiceNameResponse struct {
	BaseResponse `json:",inline"`
	ServiceName  string `json:"serviceName"`
}

// BaseWithMetricsResponse defines the base content for response DTOs (data transfer objects).
// This object and its properties correspond to the BaseWithMetricsResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/system-agent/2.1.0#/BaseWithMetricsResponse
type BaseWithMetricsResponse struct {
	BaseResponse `json:",inline"`
	ServiceName  string      `json:"serviceName"`
	Metrics      interface{} `json:"metrics"`
}

// BaseWithConfigResponse defines the base content for response DTOs (data transfer objects).
// This object and its properties correspond to the BaseWithConfigResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/system-agent/2.1.0#/BaseWithConfigResponse
type BaseWithConfigResponse struct {
	BaseResponse `json:",inline"`
	ServiceName  string      `json:"serviceName"`
	Config       interface{} `json:"config"`
}
