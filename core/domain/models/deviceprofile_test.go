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
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

var TestProfileName = "Test Profile.NAME"
var TestManufacturer = "Test Manufacturer"
var TestModel = "Test Model"
var TestProfileLabels = []string{"labe1", "label2"}
var TestProfileDescription = "Test Description"
var TestObjects = "{key1:value1, key2:value2}"
var TestProfile = DeviceProfile{DescribedObject: TestDescribedObject, Name: TestProfileName, Manufacturer: TestManufacturer, Model: TestModel, Labels: TestProfileLabels, Objects: TestObjects, DeviceResources: []DeviceObject{TestDeviceObject}, Resources: []ProfileResource{TestProfileResource}, Commands: []Command{TestCommand}}

func TestDeviceProfile_MarshalJSON(t *testing.T) {
	var emptyDeviceProfile = DeviceProfile{}
	var resultTestBytes = []byte(TestProfile.String())
	var resultEmptyTestBytes = []byte(emptyDeviceProfile.String())
	tests := []struct {
		name    string
		dp      DeviceProfile
		want    []byte
		wantErr bool
	}{
		{"successful marshal", TestProfile, resultTestBytes, false},
		{"successful empty marshal", emptyDeviceProfile, resultEmptyTestBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.dp.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("DeviceProfile.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeviceProfile.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeviceProfile_String(t *testing.T) {
	var labelSlice, _ = json.Marshal(TestProfileLabels)
	tests := []struct {
		name string
		dp   DeviceProfile
		want string
	}{
		{"device profile to string", TestProfile,
			"{\"created\":" + strconv.FormatInt(TestDescribedObject.Created, 10) +
				",\"modified\":" + strconv.FormatInt(TestDescribedObject.Modified, 10) +
				",\"origin\":" + strconv.FormatInt(TestDescribedObject.Origin, 10) +
				",\"description\":\"" + TestDescribedObject.Description + "\"" +
				",\"id\":\"\"" +
				",\"name\":\"" + TestProfileName + "\"" +
				",\"manufacturer\":\"" + TestManufacturer + "\"" +
				",\"model\":\"" + TestModel + "\"" +
				",\"labels\":" + fmt.Sprint(string(labelSlice)) +
				",\"objects\":\"" + TestObjects + "\"" +
				",\"deviceResources\":[" + TestDeviceObject.String() + "]" +
				",\"resources\":[" + TestProfileResource.String() + "]" +
				",\"commands\":[" + TestCommand.String() + "]" +
				"}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dp.String(); got != tt.want {
				t.Errorf("DeviceProfile.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
