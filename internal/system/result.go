/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package system

import "encoding/json"

const Metrics = "metrics"

// Result provides a generic interface implemented by receivers intended to return their struct as a request result.
type Result interface {
	isResult()
}

// CommonResultValue contains the fields common to all request results (regardless of whether the result indicates
// success or failure).
type CommonResultValue struct {
	Operation string `json:"operation"`
	Service   string `json:"service"`
	Executor  string `json:"executor"`
	Success   bool   `json:"Success"`
}

// SuccessResult contains the fields to be returned for a successful start/stop/restart operation request.
type SuccessResult struct {
	CommonResultValue
}

// isResult method is not called; its only purpose is to include SuccessResult in the Result abstraction.
func (r SuccessResult) isResult() {}

// metricsResultValue contains the "result" subfields specific to a metrics result.
type metricsResultValue struct {
	CpuUsedPercent float64         `json:"cpuUsedPercent"`
	MemoryUsed     int64           `json:"memoryUsed"`
	Raw            json.RawMessage `json:"raw"`
}

// MetricsSuccessResult contains the fields to be returned for a successful metrics request.
type MetricsSuccessResult struct {
	CommonResultValue
	MetricsResultValue metricsResultValue `json:"result"`
}

// isResult method is not called; its only purpose is to include MetricsSuccessResult in the Result abstraction.
func (r MetricsSuccessResult) isResult() {}

// FailureResult contains the fields to e returned for a failed request.
type FailureResult struct {
	CommonResultValue
	ErrorMessage string `json:"errorMessage"`
}

// isResult method is not called; its only purpose is to include FailureResult in the Result abstraction.
func (r FailureResult) isResult() {}

// Failure function returns a FailureResult as a Result abstraction.
func Failure(serviceName, operation, executor, errorMessage string) Result {
	return &FailureResult{
		CommonResultValue: CommonResultValue{
			Operation: operation,
			Service:   serviceName,
			Executor:  executor,
			Success:   false,
		},
		ErrorMessage: errorMessage,
	}
}

// Success function returns a SuccessResult as a Result abstraction.
func Success(serviceName, operation, executor string) Result {
	return &SuccessResult{
		CommonResultValue: CommonResultValue{
			Operation: operation,
			Service:   serviceName,
			Executor:  executor,
			Success:   true,
		},
	}
}

// MetricsSuccess function returns a MetricsSuccessResult as a Result abstraction.
func MetricsSuccess(serviceName, executor string, cpuUsedPercent float64, memoryUsed int64, raw []byte) Result {
	return &MetricsSuccessResult{
		CommonResultValue: CommonResultValue{
			Operation: Metrics,
			Service:   serviceName,
			Executor:  executor,
			Success:   true,
		},
		MetricsResultValue: metricsResultValue{
			CpuUsedPercent: cpuUsedPercent,
			MemoryUsed:     memoryUsed,
			Raw:            raw,
		},
	}
}
