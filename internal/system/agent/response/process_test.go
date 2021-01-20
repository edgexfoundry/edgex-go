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

package response

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/stretchr/testify/assert"
)

func TestProcess(t *testing.T) {
	jsonString := func() string {
		return "{\"string\":\"foo\",\"int\":123,\"float\":12.34,\"bool\":true}"
	}
	jsonMap := func() map[string]interface{} {
		return map[string]interface{}{
			"string": "foo",
			"int":    float64(123),
			"float":  12.34,
			"bool":   true,
		}
	}
	lc := logger.NewMockClient()
	tests := []struct {
		name           string
		response       string
		expectedResult map[string]interface{}
	}{
		{"invalid json", "invalidJson", map[string]interface{}{}},
		{"Empty array", "[]", map[string]interface{}{}},
		{"valid array with one json struct", "[" + jsonString() + "]", map[string]interface{}{}},
		{"valid json struct", jsonString(), jsonMap()},
		{
			"valid json struct with interior json struct",
			"{\"outside\":" + jsonString() + "}",
			map[string]interface{}{"outside": jsonMap()},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedResult, Process(test.response, lc))
		})
	}
}
