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
 * @author: Ryan Comer & Spencer Bull, Dell
 * @version: 0.5.0
 *******************************************************************************/

package models

import (
	"reflect"
	"testing"
)

func TestPropertyValue_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		pv      PropertyValue
		want    []byte
		wantErr bool
	}{
	// TODO: Add test cases.
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
	// TODO: Add test cases.
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
	type args struct {
		data []byte
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
