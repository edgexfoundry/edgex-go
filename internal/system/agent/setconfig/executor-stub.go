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

package setconfig

import (
	requests "github.com/edgexfoundry/go-mod-core-contracts/requests/configuration"
	responses "github.com/edgexfoundry/go-mod-core-contracts/responses/configuration"
)

type stubCallSet struct {
	expectedArgsSet []string                    // expected arg value for specific executor call
	outString       responses.SetConfigResponse // return value for specific executor call
}

type expectedArgsSet struct {
	service string
	sc      requests.SetConfigRequest
}

type StubSet struct {
	Called         int               // number of times stub is called
	capturedArgs   []expectedArgsSet // captures arg values for each stub call
	perCallResults stubCallSet       // expected arg value and return values for each stub call
}

func NewStubSet(results stubCallSet) StubSet {
	return StubSet{
		perCallResults: results,
	}
}

// This is a stub implementation of the SetExecutor interface.
func (m *StubSet) Do(service string, sc requests.SetConfigRequest) responses.SetConfigResponse {
	m.Called++
	m.capturedArgs = append(m.capturedArgs, expectedArgsSet{service, sc})
	return m.perCallResults.outString
}
