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

var TestDeviceService = DeviceService{Service: TestService, AdminState: "UNLOCKED"}

func TestDeviceService_MarshalJSON(t *testing.T) {
	var emptyDeviceService = DeviceService{}
	var resultTestBytes = []byte(TestDeviceService.String())
	var resultEmptyTestBytes = []byte(emptyDeviceService.String())

	tests := []struct {
		name    string
		ds      DeviceService
		want    []byte
		wantErr bool
	}{
		{"successful marshal", TestDeviceService, resultTestBytes, false},
		{"successful empty marshal", emptyDeviceService, resultEmptyTestBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ds.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("DeviceService.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeviceService.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeviceService_UnmarshalJSON(t *testing.T) {
	var resultTestBytes = []byte(TestDeviceService.String())
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		ds      *DeviceService
		args    args
		wantErr bool
	}{
		{"unmarshal normal device service with success", &TestDeviceService, args{resultTestBytes}, false},
		{"unmarshal normal device service failed", &TestDeviceService, args{[]byte("{nonsense}")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var expected = *tt.ds
			if err := tt.ds.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("DeviceService.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				// if the bytes did unmarshal, make sure they unmarshaled to correct DS by comparing it to expected results
				var unmarshaledResult = *tt.ds
				if err == nil && !reflect.DeepEqual(expected, unmarshaledResult) {
					t.Errorf("Unmarshal did not result in expected Device Service.")
				}
			}
		})
	}
}

func TestDeviceService_String(t *testing.T) {
	var labelSlice, _ = json.Marshal(TestService.Labels)
	tests := []struct {
		name string
		ds   DeviceService
		want string
	}{
		{"device service to string", TestDeviceService,
			"{\"created\":" + strconv.FormatInt(TestService.Created, 10) +
				",\"modified\":" + strconv.FormatInt(TestService.Modified, 10) +
				",\"origin\":" + strconv.FormatInt(TestService.Origin, 10) +
				",\"description\":\"" + TestDescription + "\"" +
				",\"id\":null,\"name\":\"" + TestServiceName + "\"" +
				",\"lastConnected\":" + strconv.FormatInt(TestLastConnected, 10) +
				",\"lastReported\":" + strconv.FormatInt(TestLastReported, 10) +
				",\"operatingState\":\"ENABLED\"" +
				",\"labels\":" + fmt.Sprint(string(labelSlice)) +
				",\"addressable\":" + TestAddressable.String() +
				",\"adminState\":\"UNLOCKED\"" +
				"}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ds.String(); got != tt.want {
				t.Errorf("DeviceService.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
