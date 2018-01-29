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
package enums

import (
	"reflect"
	"testing"
)

func TestGetDatabaseType(t *testing.T) {
	type args struct {
		db string
	}
	tests := []struct {
		name    string
		args    args
		want    DATABASE
		wantErr bool
	}{
		{"type is mongo", args{"mongodb"}, MONGODB, false},
		{"type is mysql", args{"mysql"}, MYSQL, false},
		{"type is unknown", args{"foo"}, INVALID, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDatabaseType(tt.args.db)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDatabaseType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDatabaseType() = %v, want %v", got, tt.want)
			}
		})
	}
}
