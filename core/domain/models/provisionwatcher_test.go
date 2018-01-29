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

var TestPWName = "TestWatcher.NAME"
var TestPWNameKey1 = "MAC"
var TestPWNameKey2 = "HTTP"
var TestPWVal1 = "00-05-1B-A1-99-99"
var TestPWVal2 = "10.0.1.1"
var TestIdentifiers = map[string]string{
	TestPWNameKey1: TestPWVal1,
	TestPWNameKey2: TestPWVal2,
}
var TestProvisionWatcher = ProvisionWatcher{BaseObject: TestBaseObject, Name: TestPWName, Identifiers: TestIdentifiers, Profile: TestProfile, Service: TestDeviceService}

func TestProvisionWatcher_MarshalJSON(t *testing.T) {
	var testPWBytes = []byte(TestProvisionWatcher.String())
	tests := []struct {
		name    string
		pw      ProvisionWatcher
		want    []byte
		wantErr bool
	}{
		{"successful marshalling", TestProvisionWatcher, testPWBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.pw.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProvisionWatcher.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProvisionWatcher.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvisionWatcher_String(t *testing.T) {
	data, _ := json.Marshal(TestIdentifiers)
	tests := []struct {
		name string
		pw   ProvisionWatcher
		want string
	}{
		{"provision watcher to string", TestProvisionWatcher,
			"{\"created\":" + strconv.FormatInt(TestProvisionWatcher.Created, 10) +
				",\"modified\":" + strconv.FormatInt(TestProvisionWatcher.Modified, 10) +
				",\"origin\":" + strconv.FormatInt(TestProvisionWatcher.Origin, 10) +
				",\"id\":\"\"" +
				",\"name\":\"" + TestPWName + "\"" +
				",\"identifiers\":" + fmt.Sprintf("%s", data) +
				",\"profile\":" + TestProvisionWatcher.Profile.String() +
				",\"service\":" + TestProvisionWatcher.Service.String() +
				",\"operatingState\":\"\"" +
				"}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pw.String(); got != tt.want {
				t.Errorf("ProvisionWatcher.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
