/*******************************************************************************
 * Copyright 2018 Dell Technologies Inc.
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
 *******************************************************************************/

package models

import "testing"

var TestBaseObject = BaseObject{Created: 123, Modified: 123}
var TestEmptyBaseObject = BaseObject{}

func TestBaseObject_String(t *testing.T) {
	tests := []struct {
		name       string
		baseObject *BaseObject
		want       string
	}{
		{"test with empty base object", &TestEmptyBaseObject, "{\"created\":0,\"modified\":0}"},
		{"test with populated base object", &TestBaseObject, "{\"created\":123,\"modified\":123}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &BaseObject{
				Created:  tt.baseObject.Created,
				Modified: tt.baseObject.Modified,
			}
			if got := o.String(); got != tt.want {
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
	var newerBaseObject = args{BaseObject{234, 234}}
	var olderBaseObject = args{BaseObject{1, 1}}
	var diffModifiedBO = args{BaseObject{123, 789}}
	tests := []struct {
		name string
		base *BaseObject
		args args
		want int
	}{
		{"test same base object", &TestBaseObject, sameBaseObject, -1},
		{"test newer base object", &TestBaseObject, newerBaseObject, 1},
		{"test older base object", &TestBaseObject, olderBaseObject, -1},
		{"test same created but different modiefied date", &TestBaseObject, diffModifiedBO, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.base.compareTo(tt.args.i); got != tt.want {
				t.Errorf("BaseObject.compareTo() = %v, want %v", got, tt.want)
			}
		})
	}
}
