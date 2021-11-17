//
// Copyright (C) 2020 IOTech Ltd
// Copyright (C) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"time"
)

// PingResponse defines the content of response content for POST Ping DTO
// This object and its properties correspond to the Ping object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-data/2.1.0#/PingResponse
type PingResponse struct {
	Versionable `json:",inline"`
	Timestamp   string `json:"timestamp"`
}

// NewPingResponse creates new PingResponse with all fields set appropriately
func NewPingResponse() PingResponse {
	return PingResponse{
		Versionable: NewVersionable(),
		Timestamp:   time.Now().Format(time.UnixDate),
	}
}
