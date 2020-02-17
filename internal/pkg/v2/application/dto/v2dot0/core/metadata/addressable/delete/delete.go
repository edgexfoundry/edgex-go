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

// delete contains v2.0 delete addressable Request and Response DTOs.
package delete

import dtoBase "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/base"

// Request defines the input for this use case.
type Request struct {
	dtoBase.Request `json:",inline"`
	ID              string `json:"id"`
}

// Response defines the output/result for this use case.
type Response struct {
	dtoBase.Response `json:",inline"`
}

// NewRequest is a factory function that returns a Request for this use case.
func NewRequest(baseRequest *dtoBase.Request, ID string) *Request {
	return &Request{
		Request: *baseRequest,
		ID:      ID,
	}
}

// NewEmptyRequest returns an uninitialized request structure for this use case.
func NewEmptyRequest() interface{} {
	var request Request
	return &request
}

// NewResponse is a factory function that returns an initialized response struct.
func NewResponse(baseResponse *dtoBase.Response) *Response {
	return &Response{
		Response: *baseResponse,
	}
}

// NewEmptyResponse returns an uninitialized response structure for this use case.
func NewEmptyResponse() interface{} {
	var response Response
	return &response
}
