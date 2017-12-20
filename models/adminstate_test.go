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

func TestAdminState_UnmarshalJSON(t *testing.T) {
	// use the referenced AdminState as the expected results
	var locked = AdminState("LOCKED")
	var unlocked = AdminState("UNLOCKED")
	var foo = AdminState("foo")
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		as      *AdminState
		args    args
		wantErr bool
	}{
		{"LOCKED marshal", &locked, args{[]byte("\"LOCKED\"")}, false},
		{"locked marshal", &locked, args{[]byte("\"locked\"")}, true},
		{"Unlocked marshal", &unlocked, args{[]byte("\"UNLOCKED\"")}, false},
		{"unlocked marshal", &unlocked, args{[]byte("\"unlocked\"")}, true},
		{"bad marshal", &foo, args{[]byte("\"goo\"")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var expected = string(*tt.as)
			if err := tt.as.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("AdminState.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				// if the bytes did unmarshal, make sure they unmarshaled to correct enum by comparing it to expected results
				var unmarshaledResult = string(*tt.as)
				if err == nil && !(IsAdminStateType(unmarshaledResult) && unmarshaledResult == expected) {
					t.Errorf("Unmarshal did not result in expected admin state string.  Expected:  %s, got: %s", expected, unmarshaledResult)
				}
			}
		})
	}
}

func TestIsAdminStateType(t *testing.T) {
	type args struct {
		as string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"LOCKED", args{"LOCKED"}, true},
		{"UNLOCKED", args{"UNLOCKED"}, true},
		{"locked", args{"locked"}, false},
		{"unlocked", args{"unlocked"}, false},
		{"non valid", args{"junk"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAdminStateType(tt.args.as); got != tt.want {
				t.Errorf("IsAdminStateType() = %v, want %v", got, tt.want)
			}
		})
	}
}
