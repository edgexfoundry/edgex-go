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

func TestOperatingState_UnmarshalJSON(t *testing.T) {
	// use the referenced Operating Stated as the expected results
	var enabled = OperatingState("ENABLED")
	var disabled = OperatingState("DISABLED")
	var foo = OperatingState("foo")

	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		os      *OperatingState
		args    args
		wantErr bool
	}{
		{"DISABLED unmarshal", &disabled, args{[]byte("\"DISABLED\"")}, false},
		{"disabled unmarshal", &disabled, args{[]byte("\"disabled\"")}, true},
		{"ENABLED unmarshal", &enabled, args{[]byte("\"ENABLED\"")}, false},
		{"enabled unmarshal", &enabled, args{[]byte("\"enabled\"")}, true},
		{"bad unmarshal", &foo, args{[]byte("\"goo\"")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var expected = string(*tt.os)
			if err := tt.os.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("OperatingState.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				// if the bytes did unmarshal, make sure they unmarshaled to correct enum by comparing it to expected results
				var unmarshaledResult = string(*tt.os)
				if err == nil && !(IsOperatingStateType(unmarshaledResult) && unmarshaledResult == expected) {
					t.Errorf("Unmarshal did not result in expected operating state string.  Expected:  %s, got: %s", expected, unmarshaledResult)
				}
			}
		})
	}
}

func TestIsOperatingStateType(t *testing.T) {
	type args struct {
		os string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"DISABLED", args{"DISABLED"}, true},
		{"ENABLED", args{"ENABLED"}, true},
		{"disabled", args{"disabled"}, false},
		{"enabled", args{"enabled"}, false},
		{"non valid", args{"junk"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsOperatingStateType(tt.args.os); got != tt.want {
				t.Errorf("IsOperatingStateType() = %v, want %v", got, tt.want)
			}
		})
	}
}
