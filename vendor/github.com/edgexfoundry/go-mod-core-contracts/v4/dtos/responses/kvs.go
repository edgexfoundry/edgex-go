//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

// MultiKVResponse defines the Response Content for GET Keys of core-keeper (GET /kvs/key/{key} API).
type MultiKVResponse struct {
	common.BaseResponse `json:",inline"`
	Response            []models.KVResponse `json:"response"`
}

// MultiKeyValueResponse defines the Response DTO for ValuesByKey HTTP client.
// This DTO is obtained from GET /kvs/key/{key} API with keyOnly is false.
type MultiKeyValueResponse struct {
	common.BaseResponse `json:",inline"`
	Response            []models.KVS `json:"response"`
}

func NewMultiKVResponse(requestId string, message string, statusCode int, resp []models.KVResponse) MultiKVResponse {
	return MultiKVResponse{
		BaseResponse: common.NewBaseResponse(requestId, message, statusCode),
		Response:     resp,
	}
}

// KeysResponse defines the Response Content for DELETE Keys controller of core-keeper (DELETE /kvs/key/{key} API).
// This DTO also defines the Response Content obtained from GET /kvs/key/{key} API with keyOnly is true.
type KeysResponse struct {
	common.BaseResponse `json:",inline"`
	Response            []models.KeyOnly `json:"response"`
}

func NewKeysResponse(requestId string, message string, statusCode int, keys []models.KeyOnly) KeysResponse {
	return KeysResponse{
		BaseResponse: common.NewBaseResponse(requestId, message, statusCode),
		Response:     keys,
	}
}
