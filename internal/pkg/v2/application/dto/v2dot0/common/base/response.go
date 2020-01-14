/*******************************************************************************
 * Copyright 2020 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package base

import (
	"encoding/json"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
)

// Response defines the Response Content for response DTOs.  This object and its properties correspond to the
// BaseResponse object in the APIv2 specification.
type Response struct {
	RequestID  string                `json:"requestId"`
	Message    interface{}           `json:"message,omitempty"`
	StatusCode infrastructure.Status `json:"statusCode"`
}

// NewResponse is a factory function that returns a Response struct.
func NewResponse(requestID string, message interface{}, statusCode infrastructure.Status) *Response {
	if message != nil {
		if _, ok := message.(string); !ok {
			marshal := func(r interface{}) string {
				b, e := json.Marshal(r)
				if e != nil {
					return e.Error()
				}
				return string(b)
			}

			// if we were passed a non-nil reference to a structure, stringify it.
			message = marshal(message)
		}
	}
	return &Response{
		RequestID:  requestID,
		Message:    message,
		StatusCode: statusCode,
	}
}

// NewResponseForSuccess is a factory function that returns a Response struct.
func NewResponseForSuccess(requestID string) *Response {
	return NewResponse(requestID, nil, infrastructure.StatusSuccess)
}
