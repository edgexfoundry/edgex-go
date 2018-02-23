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

var TestPVType = "Float"
var TestPVReadWrite = "RW"
var TestPVMinimum = "-99.99"
var TestPVMaximum = "199.99"
var TestPVDefaultValue = "0.00"
var TestPVSize = "8"
var TestPVWord = "2"
var TestPVLSB = "false"
var TestPVMask = "0x00"
var TestPVShift = "0"
var TestPVScale = "1.0"
var TestPVOffset = "0.0"
var TestPVBase = "0"
var TestPVAssertion = "0"
var TestPVSigned = true
var TestPVPrecision = "1"
var TestPropertyValue = PropertyValue{Type: TestPVType, ReadWrite: TestPVReadWrite, Minimum: TestPVMinimum, Maximum: TestPVMaximum, DefaultValue: TestPVDefaultValue, Size: TestPVSize, Word: TestPVWord, LSB: TestPVLSB, Mask: TestPVMask, Shift: TestPVShift, Scale: TestPVScale, Offset: TestPVOffset, Base: TestPVBase, Assertion: TestPVAssertion, Signed: TestPVSigned, Precision: TestPVPrecision}

func TestPropertyValue_MarshalJSON(t *testing.T) {
	var emptyPropertyValue = PropertyValue{}
	var resultTestBytes = []byte(TestPropertyValue.String())
	var emptyTestBytes = []byte(emptyPropertyValue.String())

	tests := []struct {
		name    string
		pv      PropertyValue
		want    []byte
		wantErr bool
	}{
		{"successful marshal", TestPropertyValue, resultTestBytes, false},
		{"successful empty marshal", emptyPropertyValue, emptyTestBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.pv.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("PropertyValue.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PropertyValue.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPropertyValue_String(t *testing.T) {
	tests := []struct {
		name string
		pv   PropertyValue
		want string
	}{
		{"property value to string", TestPropertyValue,
			"{\"type\":\"" + TestPVType + "\"" +
				",\"readWrite\":\"" + TestPVReadWrite + "\"" +
				",\"minimum\":\"" + TestPVMinimum + "\"" +
				",\"maximum\":\"" + TestPVMaximum + "\"" +
				",\"defaultValue\":\"" + TestPVDefaultValue + "\"" +
				",\"size\":\"" + TestPVSize + "\"" +
				",\"word\":\"" + TestPVWord + "\"" +
				",\"lsb\":\"" + TestPVLSB + "\"" +
				",\"mask\":\"" + TestPVMask + "\"" +
				",\"shift\":\"" + TestPVShift + "\"" +
				",\"scale\":\"" + TestPVScale + "\"" +
				",\"offset\":\"" + TestPVOffset + "\"" +
				",\"base\":\"" + TestPVBase + "\"" +
				",\"assertion\":\"" + TestPVAssertion + "\"" +
				",\"signed\":" + strconv.FormatBool(TestPVSigned) +
				",\"precision\":\"" + TestPVPrecision + "\"}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pv.String(); got != tt.want {
				t.Errorf("PropertyValue.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPropertyValue_UnmarshalJSON(t *testing.T) {
	var resultTestBytes = []byte(TestPropertyValue.String())

	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		p       *PropertyValue
		args    args
		wantErr bool
	}{
		{"unmarshal normal property value with success", &TestPropertyValue, args{resultTestBytes}, false},
		{"unmarshal normal property value failed", &TestPropertyValue, args{[]byte("{nonsense}")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("PropertyValue.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPropertyValue_UnmarshalYAML(t *testing.T) {
	type args struct {
		unmarshal func(interface{}) error
	}
	tests := []struct {
		name    string
		p       *PropertyValue
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.UnmarshalYAML(tt.args.unmarshal); (err != nil) != tt.wantErr {
				t.Errorf("PropertyValue.UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
