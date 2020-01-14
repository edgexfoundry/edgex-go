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

package test

// request defines the input for this use case.
type request struct {
	Message   string `json:"message"`
	DelayInMS int    `json:"delay"`
}

// response defines the output/result for this use case.
type response struct {
	Message string `json:"message"`
}

// NewRequest is a factory function that returns an initialized request struct.
func NewRequest(message string, DelayInMS int) *request {
	return &request{
		Message:   message,
		DelayInMS: DelayInMS,
	}
}

// NewEmptyRequest returns an uninitialized request structure for this use case.
func NewEmptyRequest() interface{} {
	var request request
	return &request
}

// NewResponse is a factory function that returns an initialized response struct.
func NewResponse(message string) *response {
	return &response{
		Message: message,
	}
}

// NewEmptyResponse returns an uninitialized response structure for this use case.
func NewEmptyResponse() interface{} {
	var response response
	return &response
}
