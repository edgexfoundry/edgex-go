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

var TestVDDescription = "test description"
var TestVDName = "Temperature"
var TestMin = -70
var TestMax = 140
var TestVDType = "I"
var TestUoMLabel = "C"
var TestDefaultValue = 32
var TestFormatting = "%d"
var TestVDLabels = []string{"temp", "room temp"}
var TestValueDescriptor = ValueDescriptor{Created: 123, Modified: 123, Origin: 123, Name: TestVDName, Description: TestVDDescription, Min: TestMin, Max: TestMax, DefaultValue: TestDefaultValue, Formatting: TestFormatting, Labels: TestVDLabels, UomLabel: TestUoMLabel}

func TestValueDescriptor_MarshalJSON(t *testing.T) {
	var resultTestVDBytes = []byte(TestValueDescriptor.String())
	tests := []struct {
		name    string
		v       ValueDescriptor
		want    []byte
		wantErr bool
	}{
		{"successful marshal", TestValueDescriptor, resultTestVDBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.v.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValueDescriptor.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValueDescriptor.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValueDescriptor_String(t *testing.T) {
	var labelSlice, _ = json.Marshal(TestValueDescriptor.Labels)
	tests := []struct {
		name string
		vd   ValueDescriptor
		want string
	}{
		{"value descriptor to string", TestValueDescriptor,
			"{\"id\":\"\",\"created\":" + strconv.FormatInt(TestValueDescriptor.Created, 10) +
				",\"description\":\"" + TestValueDescriptor.Description + "\"" +
				",\"modified\":" + strconv.FormatInt(TestValueDescriptor.Modified, 10) +
				",\"origin\":" + strconv.FormatInt(TestValueDescriptor.Origin, 10) +
				",\"name\":\"" + TestValueDescriptor.Name + "\"" +
				",\"min\":" + strconv.Itoa(TestValueDescriptor.Min.(int)) +
				",\"max\":" + strconv.Itoa(TestValueDescriptor.Max.(int)) +
				",\"defaultValue\":" + strconv.Itoa(TestValueDescriptor.DefaultValue.(int)) +
				",\"type\":null" +
				",\"uomLabel\":\"" + TestValueDescriptor.UomLabel + "\"" +
				",\"formatting\":\"" + TestValueDescriptor.Formatting + "\"" +
				",\"labels\":" + fmt.Sprint(string(labelSlice)) + "}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.vd.String(); got != tt.want {
				t.Errorf("ValueDescriptor.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
