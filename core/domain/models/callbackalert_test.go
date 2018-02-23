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

var TestCallbackAlert = CallbackAlert{"DEVICE", "1234"}

func TestCallbackAlert_MarshalJSON(t *testing.T) {
	var testNoIDCallbackAlert = CallbackAlert{"DEVICE", ""}
	var testCallbackAlertBytes = []byte(TestCallbackAlert.String())
	var testNoIDCallbackAlertBytes = []byte(testNoIDCallbackAlert.String())
	tests := []struct {
		name    string
		ca      CallbackAlert
		want    []byte
		wantErr bool
	}{
		{"successful marshal of callback", TestCallbackAlert, testCallbackAlertBytes, false},
		{"successful marshal of no id callback", testNoIDCallbackAlert, testNoIDCallbackAlertBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ca.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("CallbackAlert.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CallbackAlert.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCallbackAlert_String(t *testing.T) {
	tests := []struct {
		name string
		ca   CallbackAlert
		want string
	}{
		{"callback alert to string", TestCallbackAlert, "{\"type\":\"DEVICE\",\"id\":\"1234\"}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ca.String(); got != tt.want {
				t.Errorf("CallbackAlert.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
