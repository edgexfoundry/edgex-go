//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package base

// Response defines the Response Content for response DTOs.  This object and its properties correspond to the
// BaseResponse object in the APIv2 specification.
type Response struct {
	CorrelationID string      `json:"correlationId"`
	RequestID     string      `json:"requestId"`
	Message       interface{} `json:"message,omitempty"`
	StatusCode    int         `json:"statusCode"`
}
