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

var TestPut = Put{Action: TestAction, ParameterNames: TestExpectedvalues}

func TestPut_MarshalJSON(t *testing.T) {
	var testPutBytes = []byte(TestPut.String())
	tests := []struct {
		name    string
		p       Put
		want    []byte
		wantErr bool
	}{
		{"successful marshalling", TestPut, testPutBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Put.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Put.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPut_String(t *testing.T) {
	tests := []struct {
		name string
		p    Put
		want string
	}{
		{"put to string", TestPut,
			"{\"path\":\"" + TestActionPath +
				"\",\"responses\":[{\"code\":\"" + TestCode +
				"\",\"description\":\"" + TestDescription +
				"\",\"expectedValues\":[\"" + TestExpectedvalue1 +
				"\",\"" + TestExpectedvalue2 +
				"\"]}],\"parameterNames\":[\"" + TestExpectedvalue1 + "\",\"" + TestExpectedvalue2 + "\"]}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.String(); got != tt.want {
				t.Errorf("Put.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPut_AllAssociatedValueDescriptors(t *testing.T) {
	var testMap = make(map[string]string)
	type args struct {
		vdNames *map[string]string
	}
	tests := []struct {
		name string
		p    *Put
		args args
	}{
		{"put assoc val descs", &TestPut, args{vdNames: &testMap}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.AllAssociatedValueDescriptors(tt.args.vdNames)
			if len(*tt.args.vdNames) != 2 {
				t.Error("Associated value descriptor size > than expected")
			}
		})
	}
}
