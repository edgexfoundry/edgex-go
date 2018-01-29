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
	"strconv"
	"testing"
)

var TestValueDescriptorName = "Temperature"
var TestValue = "45"
var TestReading = Reading{Pushed: 123, Created: 123, Origin: 123, Modified: 123, Device: TestDeviceName, Name: TestValueDescriptorName, Value: TestValue}

func TestReading_MarshalJSON(t *testing.T) {
	var emptyReading = Reading{}
	var resultTestBytes = []byte(TestReading.String())
	var resultEmptyTestBytes = []byte(emptyReading.String())

	tests := []struct {
		name    string
		r       Reading
		want    []byte
		wantErr bool
	}{
		{"successful marshal", TestReading, resultTestBytes, false},
		{"successful empty marshal", emptyReading, resultEmptyTestBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Reading.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reading.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReading_String(t *testing.T) {
	tests := []struct {
		name string
		r    Reading
		want string
	}{
		{"reading to string", TestReading,
			"{\"id\":\"\"" +
				",\"pushed\":" + strconv.FormatInt(TestReading.Pushed, 10) +
				",\"created\":" + strconv.FormatInt(TestReading.Created, 10) +
				",\"origin\":" + strconv.FormatInt(TestReading.Origin, 10) +
				",\"modified\":" + strconv.FormatInt(TestReading.Modified, 10) +
				",\"device\":\"" + TestDeviceName + "\"" +
				",\"name\":\"" + TestValueDescriptorName + "\"" +
				",\"value\":\"" + TestValue + "\"" +
				"}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.String(); got != tt.want {
				t.Errorf("Reading.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
