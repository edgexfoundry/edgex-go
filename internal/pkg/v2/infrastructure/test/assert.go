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

package test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/correlationid"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	NoDelayDescription           = "no delay"
	StandardDelayInMS            = 75
	StandardDelayInMSDescription = "75ms delay"
	standardDeviationInMS        = 50
	None                         = "none"
	One                          = "one"
	Two                          = "two"
	TypeValid                    = "Valid"
	TypeEmpty                    = "Empty"
	TypeNoRoute                  = "NoRoute"
	TypeCannotUnmarshal          = "CannotUnmarshal"
	TypeEmptyRequestId           = "EmptyRequestId"
	TypeUpdateNonExistent        = "UpdateNonExistent"
	TypeUpdateOneProperty        = "UpdateOneProperty"
	TypeUpdateAllProperties      = "UpdateAllProperties"
)

// assertHeader provides common implementation to assert an HTTP header has an expected value.
func assertHeader(t *testing.T, headers http.Header, headerKey, expectedValue string) {
	value, ok := headers[headerKey]
	_ = assert.True(t, ok) &&
		assert.Len(t, value, 1) &&
		assert.Equal(t, expectedValue, value[0])
}

// AssertCorrelationID provides common implementation to assert HTTP correlationID header has expected value.
func AssertCorrelationID(t *testing.T, headers http.Header, expectedValue string) {
	assertHeader(t, headers, correlationid.HTTPHeader, expectedValue)
}

// assertContentType provides common implementation to assert HTTP content-type header has a single expected value.
func assertContentType(t *testing.T, headers http.Header, expectedType string) {
	assertHeader(t, headers, "Content-Type", expectedType)
}

// AssertContentTypeIsJSON provides common implementation to assert HTTP content-type header indicates JSON.
func AssertContentTypeIsJSON(t *testing.T, headers http.Header) {
	assertContentType(t, headers, "application/json")
}

// assertJSONBodyEquals provides common implementation to assert a response body has the expected value.
func assertJSONBodyEquals(t *testing.T, expected interface{}, actual []byte) {
	_ = assert.Equal(t, string(Marshal(t, expected)), string(actual))
}

// permutation is recursive function used by assertJSONBodyEqualsForMultiple to test all permutations of a.
func permutation(a []interface{}, f func(a []interface{}) bool, i int) bool {
	if i > len(a) {
		return f(a)
	}
	if permutation(a, f, i+1) {
		return true
	}
	for j := i + 1; j < len(a); j++ {
		a[i], a[j] = a[j], a[i]
		if permutation(a, f, i+1) {
			return true
		}
		a[i], a[j] = a[j], a[i]
	}
	return false
}

// assertJSONBodyEqualsForMultiple provides common implementation to assert a response body has the expected values
// in any order; used for endpoint tests that send multiple requests and expect multiple responses.
func assertJSONBodyEqualsForMultiple(t *testing.T, expected []interface{}, actual []byte) {
	var actualObject interface{}
	Unmarshal(t, actual, &actualObject)
	_ = assert.True(
		t,
		permutation(
			expected,
			func(a []interface{}) bool {
				var expectedObject interface{}
				Unmarshal(t, Marshal(t, expected), &expectedObject)
				return reflect.DeepEqual(expectedObject, actualObject)
			},
			0,
		),
		"unable to match any variation of expected (%s) against actual (%s)",
		string(Marshal(t, expected)),
		string(actual),
	)
}

// AssertJSONBody provides common implementation to assert a response body has the expected values regardless of
// whether the response is a single object or an array of objects.
func AssertJSONBody(t *testing.T, expected interface{}, actual []byte) {
	if array, ok := expected.([]interface{}); ok {
		assertJSONBodyEqualsForMultiple(t, array, actual)
	} else {
		assertJSONBodyEquals(t, expected, actual)
	}
}

// AssertElapsedInsideDeviation provides a common implementation to assert a timer falls within expected range.
func AssertElapsedInsideDeviation(t *testing.T, timer *Timer, expectedInMS, deviationInMS int) {
	assert.True(t, timer.insideDeviation(expectedInMS, deviationInMS))
}

// AssertElapsedInsideStandardDeviation provides a common implementation to assert a timer falls within expected range.
func AssertElapsedInsideStandardDeviation(t *testing.T, timer *Timer, expectedInMS int) {
	AssertElapsedInsideDeviation(t, timer, expectedInMS, standardDeviationInMS)
}

// AssertIsIdentity provides a common implementation to assert identity is a UUID.
func AssertIsIdentity(t *testing.T, identity string) {
	_, err := uuid.Parse(identity)
	assert.Nil(t, err)
}
