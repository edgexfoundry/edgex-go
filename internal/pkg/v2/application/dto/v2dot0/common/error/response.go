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

package error

import dtoBase "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/base"

// Response defines the content of the generic error DTO.  This object and its properties correspond to the
// ErrorResponse object in the APIv2 specification.
type Response struct {
	dtoBase.Response `json:",inline"`
}

// NewResponse is a factory function that returns an initialized Response struct.
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
