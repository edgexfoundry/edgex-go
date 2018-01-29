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

var TestUnitsType = "String"
var TestUnitsRW = "R"
var TestUnitsDV = "Degrees Fahrenheit"
var TestUnits = Units{Type: TestUnitsType, ReadWrite: TestUnitsRW, DefaultValue: TestUnitsDV}

func TestUnits_MarshalJSON(t *testing.T) {
	var testUnitsBytes = []byte(TestUnits.String())
	tests := []struct {
		name    string
		u       Units
		want    []byte
		wantErr bool
	}{
		{"successful marshalling", TestUnits, testUnitsBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.u.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Units.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Units.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnits_String(t *testing.T) {
	tests := []struct {
		name string
		u    Units
		want string
	}{
		{"units to string", TestUnits,
			"{\"type\":\"" + TestUnitsType + "\"" +
				",\"readWrite\":\"" + TestUnitsRW + "\"" +
				",\"defaultValue\":\"" + TestUnitsDV + "\"}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.String(); got != tt.want {
				t.Errorf("Units.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
