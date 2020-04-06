//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package base

// Request defines the Request Content for request DTOs. This object and its properties correspond to the BaseRequest
// object in the APIv2 specification.
type Request struct {
	CorrelationID string `json:"correlationId"`
	RequestID     string `json:"requestId"`
}
