//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
)

// KeyDataResponse defines the Response Content for GET KeyData DTOs.
type KeyDataResponse struct {
	dtoCommon.BaseResponse `json:",inline"`
	KeyData                dtos.KeyData `json:"keyData"`
}

func NewKeyDataResponse(requestId string, message string, statusCode int, keyData dtos.KeyData) KeyDataResponse {
	return KeyDataResponse{
		BaseResponse: dtoCommon.NewBaseResponse(requestId, message, statusCode),
		KeyData:      keyData,
	}
}
