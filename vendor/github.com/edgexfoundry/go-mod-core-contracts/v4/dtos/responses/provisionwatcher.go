//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
)

// ProvisionWatcherResponse defines the Response Content for GET ProvisionWatcher DTOs.
type ProvisionWatcherResponse struct {
	common.BaseResponse `json:",inline"`
	ProvisionWatcher    dtos.ProvisionWatcher `json:"provisionWatcher"`
}

func NewProvisionWatcherResponse(requestId string, message string, statusCode int, pw dtos.ProvisionWatcher) ProvisionWatcherResponse {
	return ProvisionWatcherResponse{
		BaseResponse:     common.NewBaseResponse(requestId, message, statusCode),
		ProvisionWatcher: pw,
	}
}

// MultiProvisionWatchersResponse defines the Response Content for GET multiple ProvisionWatcher DTOs.
type MultiProvisionWatchersResponse struct {
	common.BaseWithTotalCountResponse `json:",inline"`
	ProvisionWatchers                 []dtos.ProvisionWatcher `json:"provisionWatchers"`
}

func NewMultiProvisionWatchersResponse(requestId string, message string, statusCode int, totalCount uint32, pws []dtos.ProvisionWatcher) MultiProvisionWatchersResponse {
	return MultiProvisionWatchersResponse{
		BaseWithTotalCountResponse: common.NewBaseWithTotalCountResponse(requestId, message, statusCode, totalCount),
		ProvisionWatchers:          pws,
	}
}
