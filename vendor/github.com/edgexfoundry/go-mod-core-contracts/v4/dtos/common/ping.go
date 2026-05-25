//
// Copyright (C) 2020 IOTech Ltd
// Copyright (C) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"time"
)

// PingResponse defines the content of response content for GET Ping DTO
type PingResponse struct {
	Versionable `json:",inline"`
	Timestamp   string `json:"timestamp"`
	ServiceName string `json:"serviceName"`
}

// NewPingResponse creates new PingResponse with all fields set appropriately
func NewPingResponse(serviceName string) PingResponse {
	return PingResponse{
		Versionable: NewVersionable(),
		Timestamp:   time.Now().Format(time.UnixDate),
		ServiceName: serviceName,
	}
}
