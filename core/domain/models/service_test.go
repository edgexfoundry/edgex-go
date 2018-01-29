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

var TestServiceName = "test service"
var TestService = Service{DescribedObject: TestDescribedObject, Name: TestServiceName, LastConnected: TestLastConnected, LastReported: TestLastReported, OperatingState: "ENABLED", Labels: TestLabels, Addressable: TestAddressable}

func TestService_MarshalJSON(t *testing.T) {
	var resultTestServiceBytes = []byte(TestService.String())
	tests := []struct {
		name    string
		s       Service
		want    []byte
		wantErr bool
	}{
		{"successful marshal", TestService, resultTestServiceBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_String(t *testing.T) {
	var labelSlice, _ = json.Marshal(TestService.Labels)
	tests := []struct {
		name string
		dp   Service
		want string
	}{
		{"service to string", TestService,
			"{\"created\":" + strconv.FormatInt(TestService.Created, 10) +
				",\"modified\":" + strconv.FormatInt(TestService.Modified, 10) +
				",\"origin\":" + strconv.FormatInt(TestService.Origin, 10) +
				",\"description\":\"" + TestDescription + "\"" +
				",\"id\":null,\"name\":\"" + TestServiceName + "\"" +
				",\"lastConnected\":" + strconv.FormatInt(TestLastConnected, 10) +
				",\"lastReported\":" + strconv.FormatInt(TestLastReported, 10) +
				",\"operatingState\":\"ENABLED\"" +
				",\"labels\":" + fmt.Sprint(string(labelSlice)) +
				",\"addressable\":" + TestAddressable.String() + "}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dp.String(); got != tt.want {
				t.Errorf("Service.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
