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

var TestDeviceName = "test device name"
var TestLabels = []string{"MODBUS", "TEMP"}
var TestLastConnected = int64(1000000)
var TestLastReported = int64(1000000)
var TestLocation = "{40lat;45long}"
var TestDevice = Device{DescribedObject: TestDescribedObject, Name: TestDeviceName, AdminState: "UNLOCKED", OperatingState: "ENABLED", Addressable: TestAddressable, LastReported: TestLastReported, LastConnected: TestLastConnected, Labels: TestLabels, Location: TestLocation, Service: TestDeviceService, Profile: TestProfile}

func TestDevice_MarshalJSON(t *testing.T) {
	var testDeviceBytes = []byte(TestDevice.String())

	tests := []struct {
		name    string
		d       Device
		want    []byte
		wantErr bool
	}{
		{"successful marshal", TestDevice, testDeviceBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Device.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Device.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDevice_String(t *testing.T) {
	var labelSlice, _ = json.Marshal(TestDevice.Labels)
	tests := []struct {
		name string
		d    Device
		want string
	}{
		{"device to string", TestDevice,
			"{\"created\":" + strconv.FormatInt(TestDevice.Created, 10) +
				",\"modified\":" + strconv.FormatInt(TestDevice.Modified, 10) +
				",\"origin\":" + strconv.FormatInt(TestDevice.Origin, 10) +
				",\"description\":\"" + TestDescription + "\"" +
				",\"id\":null,\"name\":\"" + TestDevice.Name + "\"" +
				",\"adminState\":\"UNLOCKED\",\"operatingState\":\"ENABLED\",\"addressable\":" + TestAddressable.String() +
				",\"lastConnected\":" + strconv.FormatInt(TestLastConnected, 10) +
				",\"lastReported\":" + strconv.FormatInt(TestLastReported, 10) +
				",\"labels\":" + fmt.Sprint(string(labelSlice)) +
				",\"location\":\"" + TestLocation + "\"" +
				",\"service\":" + TestDevice.Service.String() +
				",\"profile\":" + TestDevice.Profile.String() +
				"}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.String(); got != tt.want {
				t.Errorf("Device.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDevice_AllAssociatedValueDescriptors(t *testing.T) {
	var assocVD []string
	type args struct {
		vdNames *[]string
	}
	tests := []struct {
		name string
		d    *Device
		args args
	}{
		{"get associated value descriptors", &TestDevice, args{vdNames: &assocVD}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.d.AllAssociatedValueDescriptors(tt.args.vdNames)
			if len(*tt.args.vdNames) != 2 {
				t.Error("Associated value descriptor size > than expected")
			}
		})
	}
}
