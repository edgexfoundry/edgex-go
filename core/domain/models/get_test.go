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

import (
	"reflect"
	"testing"
)

var TestGet = Get{Action: TestAction}

func TestGet_MarshalJSON(t *testing.T) {
	var testGetBytes = []byte(TestGet.String())
	tests := []struct {
		name    string
		g       Get
		want    []byte
		wantErr bool
	}{
		{"successful marshalling", TestGet, testGetBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.g.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Get.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGet_String(t *testing.T) {
	tests := []struct {
		name string
		g    Get
		want string
	}{
		{"get to string", TestGet, TestGet.Action.String()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.String(); got != tt.want {
				t.Errorf("Get.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGet_AllAssociatedValueDescriptors(t *testing.T) {
	var testMap = make(map[string]string)
	type args struct {
		vdNames *map[string]string
	}
	tests := []struct {
		name string
		g    *Get
		args args
	}{
		{"get assoc val descs", &TestGet, args{vdNames: &testMap}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.g.AllAssociatedValueDescriptors(tt.args.vdNames)
			if len(*tt.args.vdNames) != 2 {
				t.Error("Associated value descriptor size > than expected")
			}
		})
	}
}
