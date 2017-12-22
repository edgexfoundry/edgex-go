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

var TestProfileResourceName = "test profile resource name"
var TestProfileResource = ProfileResource{Name: TestProfileResourceName, Get: []ResourceOperation{TestResourceOperation}, Set: []ResourceOperation{TestResourceOperation}}

func TestProfileResource_MarshalJSON(t *testing.T) {
	var emptyProfileResource = ProfileResource{}
	var resultTestBytes = []byte(TestProfileResource.String())
	var emptyTestBytes = []byte(emptyProfileResource.String())
	tests := []struct {
		name    string
		pr      ProfileResource
		want    []byte
		wantErr bool
	}{
		{"successful marshal", TestProfileResource, resultTestBytes, false},
		{"successful empty marshal", emptyProfileResource, emptyTestBytes, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.pr.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProfileResource.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProfileResource.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProfileResource_String(t *testing.T) {
	tests := []struct {
		name string
		pr   ProfileResource
		want string
	}{
		{"profile resource to string", TestProfileResource,
			"{\"name\":\"" + TestProfileResourceName + "\"" +
				",\"get\":[" + TestResourceOperation.String() +
				"],\"set\":[" + TestResourceOperation.String() + "]}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pr.String(); got != tt.want {
				t.Errorf("ProfileResource.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
