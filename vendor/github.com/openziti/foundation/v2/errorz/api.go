/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package errorz

import "net/http"

type ApiError struct {
	AppCode     string      `json:"code"`
	Message     string      `json:"message"`
	Status      int         `json:"-"`
	Cause       error       `json:"cause"`
	AppendCause bool        `json:"-"`
	Headers     http.Header `json:"-"`
}

func (e ApiError) Error() string {
	s := e.AppCode + ": " + e.Message

	if e.Cause != nil && e.AppendCause {
		s = s + ": " + e.Cause.Error()
	}
	return s
}

func (e ApiError) Code() int32 {
	return int32(e.Status)
}
