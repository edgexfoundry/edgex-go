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

package batchdto

import "encoding/json"

// response defines the content of the common envelope encapsulating every response.
type response struct {
	Common  `json:",inline"`
	Content interface{} `json:"content"`
}

// NewResponse is a factory function that returns an initialized response.
func NewResponse(common *Common, content interface{}) *response {
	return &response{
		Common:  *common,
		Content: content,
	}
}

// NewResponseFromRequest is a factory function that returns an initialized response struct; Version, Kind, and Action
// fields are taken from provided request.
func NewResponseFromRequest(request *request, content interface{}) *response {
	return &response{
		Common:  *NewResponseCommonFromRequestCommon(&request.Common),
		Content: content,
	}
}

// NewEmptyResponse returns an uninitialized common envelope structure.
func NewEmptyResponse() interface{} {
	var response response
	return &response
}

// TestResponse defines the content of the common envelope encapsulating every response; type facilitates acceptance
// testing.
type TestResponse struct {
	Common  `json:",inline"`
	Content *json.RawMessage `json:"content"`
}

// EmptyTestResponseSlice is a factory function that returns an empty TestResponse slice.
func EmptyTestResponseSlice() []TestResponse {
	return []TestResponse{}
}

// NewTestResponse is a factory function that returns a TestResponse struct.
func NewTestResponse(common *Common, content *json.RawMessage) *TestResponse {
	return &TestResponse{
		Common:  *common,
		Content: content,
	}
}
