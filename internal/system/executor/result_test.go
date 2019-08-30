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

package executor

import "testing"

// TestIsResultForCoverage ensures test coverage by calling MetricsResultValue.isResult implementations; these are
// never called from EdgeX code as their only purpose is to support a Golang union equivalent (i.e. to return
// an abstract result struct whose content varies).
func TestIsResultForCoverage(t *testing.T) {
	tests := []struct {
		name string
		sut  Result
	}{
		{"SuccessResult", SuccessResult{}},
		{"MetricsSuccessResult", MetricsSuccessResult{}},
		{"FailureResult", FailureResult{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.sut.isResult()
		})
	}
}
