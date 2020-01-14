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

// Request defines the Request Content for request DTOs. This object and its properties correspond to the BaseRequest
// object in the APIv2 specification.
type Request struct {
	RequestID string `json:"requestId"`
}

// NewRequest is a factory function that returns a Request struct.
func NewRequest(requestID string) *Request {
	return &Request{
		RequestID: requestID,
	}
}
