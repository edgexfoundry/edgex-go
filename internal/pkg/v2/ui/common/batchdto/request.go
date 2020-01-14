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

// request defines the Content of the Common envelope encapsulating every request.
type request struct {
	Common  `json:",inline"`
	Content *json.RawMessage `json:"content"`
}

// EmptyRequestSlice is a factory function that returns an empty request slice.  This keeps request implementation
// private.
func EmptyRequestSlice() []request {
	return []request{}
}

// TestRequest defines the content of the common envelope encapsulating every request; type facilitates acceptance
// testing.
type TestRequest struct {
	Common  `json:",inline"`
	Content interface{} `json:"content"`
}

// NewTestRequest is a factory function that returns a TestRequest struct.
func NewTestRequest(common *Common, content interface{}) *TestRequest {
	return &TestRequest{
		Common:  *common,
		Content: content,
	}
}
