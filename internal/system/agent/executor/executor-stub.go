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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type stubCall struct {
	expectedArgs []string // expected arg value for specific executor call
	outString    string   // return value for specific executor call
	outError     error    // return value for specific executor call
}

type Stub struct {
	Called         int        // number of times stub is called
	capturedArgs   [][]string // captures arg values for each stub call
	perCallResults []stubCall // expected arg value and return values for each stub call
}

func NewStub(results []stubCall) Stub {
	return Stub{
		perCallResults: results,
	}
}

// CommandExecutor provides the common callout to the configuration-defined executor.  This is a stub implementation of
// the CommandExecutor interface.
func (m *Stub) CommandExecutor(executorPath, serviceName, operation string) (string, error) {
	m.Called++
	m.capturedArgs = append(m.capturedArgs, []string{executorPath, serviceName, operation})
	return m.perCallResults[m.Called-1].outString, m.perCallResults[m.Called-1].outError
}

// AssertArgsAreEqual
func assertArgsAreEqual(t *testing.T, expected []string, actual []string) {
	assert.Equal(t, len(expected), len(actual))
	for key, expectedValue := range expected {
		assert.Equal(t, expectedValue, actual[key])
	}
}
