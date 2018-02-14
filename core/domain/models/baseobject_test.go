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

import "testing"

var TestBaseObject = BaseObject{Created: 123, Modified: 123, Origin: 123}
var EmptyBaseObject = BaseObject{}

func TestBaseObject_String(t *testing.T) {
	tests := []struct {
		name       string
		baseObject *BaseObject
		want       string
	}{
		{"empty base", &EmptyBaseObject, "{\"created\":0,\"modified\":0,\"origin\":0}"},
		{"populated base", &TestBaseObject, "{\"created\":123,\"modified\":123,\"origin\":123}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.baseObject.String(); got != tt.want {
				t.Errorf("BaseObject.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseObject_compareTo(t *testing.T) {
	type args struct {
		i BaseObject
	}
	var sameBaseObject = args{TestBaseObject}
	var newerBaseObject = args{BaseObject{234, 234, 234}}
	var olderBaseObject = args{BaseObject{1, 1, 1}}
	tests := []struct {
		name string
		ba   *BaseObject
		args args
		want int
	}{
		{"same object", &TestBaseObject, sameBaseObject, -1},
		{"newer", &TestBaseObject, newerBaseObject, 1},
		{"older", &TestBaseObject, olderBaseObject, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ba.compareTo(tt.args.i); got != tt.want {
				t.Errorf("BaseObject.compareTo() = %v, want %v", got, tt.want)
			}
		})
	}
}
