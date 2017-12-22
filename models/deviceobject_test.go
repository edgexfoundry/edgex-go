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

var TestDeviceObjectDescription = "test device object description"
var TestDeviceObjectName = "test device object name"
var TestDeviceObjectTag = "test device object tag"
var TestDeviceObject = DeviceObject{Description: TestDeviceObjectDescription, Name: TestDeviceObjectName, Tag: TestDeviceObjectTag, Properties: TestProfileProperty}

func TestDeviceObject_MarshalJSON(t *testing.T) {
	var emptyDeviceObject = DeviceObject{}
	var resultTestBytes = []byte(TestDeviceObject.String())
	var resultEmptyTestBytes = []byte(emptyDeviceObject.String())
	tests := []struct {
		name    string
		do      DeviceObject
		want    []byte
		wantErr bool
	}{
		{"successful marshal", TestDeviceObject, resultTestBytes, false},
		{"successful empty marshal", emptyDeviceObject, resultEmptyTestBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.do.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("DeviceObject.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeviceObject.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeviceObject_String(t *testing.T) {
	tests := []struct {
		name string
		do   DeviceObject
		want string
	}{
		{"device object to string", TestDeviceObject,
			"{\"description\":\"" + TestDeviceObjectDescription + "\"" +
				",\"name\":\"" + TestDeviceObjectName + "\"" +
				",\"tag\":\"" + TestDeviceObjectTag + "\"" +
				",\"properties\":" + TestProfileProperty.String() +
				",\"attributes\":null}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.do.String(); got != tt.want {
				t.Errorf("DeviceObject.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
