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
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type stubCall struct {
	expectedArgs []string // expected arg value for specific executor call
	outString    string   // return value for specific executor call
	outError     error    // return value for specific executor call
}

func argsToString(a []string) string {
	return strings.Join(a, "_")
}

type Stub struct {
	mutex          sync.Mutex
	Called         int                 // number of times stub is called
	capturedArgs   [][]string          // captures arg values for each stub call
	perCallResults map[string]stubCall // expected arg value and return values for each stub call
}

func NewStub(results map[string]stubCall) Stub {
	return Stub{
		perCallResults: results,
	}
}

// CommandExecutor provides the common callout to the configuration-defined executor.  This is a stub implementation of
// the CommandExecutor interface.
func (m *Stub) CommandExecutor(executorPath, serviceName, operation string) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Called++
	m.capturedArgs = append(m.capturedArgs, []string{executorPath, serviceName, operation})
	if _, ok := m.perCallResults[serviceName]; ok {
		return m.perCallResults[serviceName].outString, m.perCallResults[serviceName].outError
	}
	return "serviceName not found", nil
}

// assertArgsAreEqualInAnyOrder compares expected to actual to ensure they are the same; note order may vary.
func assertArgsAreEqualInAnyOrder(t *testing.T, expected []string, actual []string) {
	actualLen := len(actual)
	assert.Equal(t, len(expected), actualLen)

	diff := make(map[string]int, actualLen)
	for actualKey := range actual {
		diff[actual[actualKey]]++
	}

	for _, expectedValue := range expected {
		if _, ok := diff[expectedValue]; !ok {
			assert.Fail(t, "missing %s", expectedValue)
			return
		}
		diff[expectedValue]--
		if diff[expectedValue] == 0 {
			delete(diff, expectedValue)
		}
	}

	if len(diff) != 0 {
		assert.Fail(t, "received unexpectedly %v", diff)
	}
}

// assertResultsAreEqualInAnyOrder compares expected to actual to ensure they are the same; note order may vary
func assertResultsAreEqualInAnyOrder(t *testing.T, expected []interface{}, actual []interface{}) {
	convertSlice := func(a []interface{}) []string {
		interfaceToString := func(i interface{}) string {
			b, e := json.Marshal(i)
			if e != nil {
				assert.Fail(t, "failure %v attempting to marshal %v", e, i)
			}
			return string(b)
		}

		r := []string{}
		for k := range a {
			r = append(r, interfaceToString(a[k]))
		}
		return r
	}

	assertArgsAreEqualInAnyOrder(t, convertSlice(expected), convertSlice(actual))
}
