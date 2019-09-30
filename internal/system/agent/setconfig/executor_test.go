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
	"testing"

	requests "github.com/edgexfoundry/go-mod-core-contracts/requests/configuration"
	responses "github.com/edgexfoundry/go-mod-core-contracts/responses/configuration"

	"github.com/stretchr/testify/assert"
)

func TestSetExecutorWithNoServices(t *testing.T) {
	executor := NewStubSet(stubCallSet{})
	sut := New(&executor)
	sc := requests.SetConfigRequest{}
	actual := sut.Do([]string{}, sc)

	assert.Equal(t, resultType{Configuration: resultConfigurationType{}}, actual)
	assert.Equal(t, executor.Called, 0)
}

func TestSetExecutorWithServices(t *testing.T) {
	const serviceName = "serviceName"
	expectedResult := resultType{
		Configuration: resultConfigurationType{
			"serviceName": {Success: false, Description: ""},
		}}

	sc := requests.SetConfigRequest{Key: "Writable.LogLevel", Value: "INFO"}

	tests := []struct {
		name           string
		services       []string
		expectedResult resultType
		executorCalls  stubCallSet
	}{
		{
			"one service is the target of the set operation",
			[]string{serviceName},
			expectedResult,
			stubCallSet{[]string{serviceName}, responses.SetConfigResponse{}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := NewStubSet(test.executorCalls)
			sut := New(&executor)
			actualResult := sut.Do(test.services, sc)
			assert.Equal(t, test.expectedResult, actualResult)
		})
	}
}
