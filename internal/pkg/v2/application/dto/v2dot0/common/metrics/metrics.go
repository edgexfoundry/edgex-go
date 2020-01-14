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

// metrics contains v2.0 metrics request and response DTOs.
package metrics

import dtoBase "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/base"

// Request defines the input for this use case. This object and its properties correspond to the MetricsRequest object
// in the APIv2 specification.
type Request struct {
	dtoBase.Request `json:",inline"`
}

// Response defines the output/result for this use case. This object and its properties correspond to the
// MetricsResponse object in the APIv2 specification.
type Response struct {
	dtoBase.Response `json:",inline"`
	Alloc            uint64  `json:"memAlloc"`
	TotalAlloc       uint64  `json:"memTotalAlloc"`
	Sys              uint64  `json:"memSys"`
	Mallocs          uint64  `json:"memMallocs"`
	Frees            uint64  `json:"memFrees"`
	LiveObjects      uint64  `json:"memLiveObjects"`
	CpuBusyAvg       float64 `json:"cpuBusyAvg"`
}

// NewRequest is a factory function that returns a Request for this use case.
func NewRequest(baseRequest *dtoBase.Request) interface{} {
	return &Request{
		Request: *baseRequest,
	}
}

// NewEmptyRequest returns an uninitialized request structure for this use case.
func NewEmptyRequest() interface{} {
	var request Request
	return &request
}

// NewResponse is a factory function that returns an initialized Response struct.
func NewResponse(baseResponse dtoBase.Response, alloc, totalAlloc, sys, mallocs, frees, liveObjects uint64, cpuBusyAvg float64) *Response {
	return &Response{
		Response:    baseResponse,
		Alloc:       alloc,
		TotalAlloc:  totalAlloc,
		Sys:         sys,
		Mallocs:     mallocs,
		Frees:       frees,
		LiveObjects: liveObjects,
		CpuBusyAvg:  cpuBusyAvg,
	}
}

// NewEmptyResponse returns an uninitialized response structure for this use case.
func NewEmptyResponse() interface{} {
	var response Response
	return &response
}
