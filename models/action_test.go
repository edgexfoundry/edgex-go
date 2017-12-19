/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
 *
 * @microservice: core-domain-go library
 * @author: Jim White, Dell
 * @version: 0.5.0
 *******************************************************************************/

package models

import "testing"

const testCode = "200"
const testDescription = "ok"
const testExpectedvalue1 = "temperature"
const testExpectedvalue2 = "humidity"
const testActionPath = "test/path"
const testFooBar = "foobar"

var TestAction = Action{testActionPath, []Response{Response{testCode, testDescription, []string{testExpectedvalue1, testExpectedvalue2}}}}
var EmptyAction = Action{}

func TestAction_String(t *testing.T) {

	tests := []struct {
		name   string
		action Action
		want   string
	}{
		{"full action", TestAction, "{\"path\":\"" + testActionPath + "\",\"responses\":[{\"code\":\"" + testCode + "\",\"description\":\"" + testDescription + "\",\"expectedValues\":[\"" + testExpectedvalue1 + "\",\"" + testExpectedvalue2 + "\"]}]}"},
		{"empty action", EmptyAction, "{\"path\":\"\",\"responses\":null}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.action.String(); got != tt.want {
				t.Errorf("Action.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
