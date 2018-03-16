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
	"testing"
)

func TestGetDatabaseType(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    DATABASE
		wantErr bool
	}{
		{"type is mongo", "mongodb", MONGODB, false},
		{"type is mysql", "mysql", MYSQL, false},
		{"type is unknown", "foo", INVALID, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDatabaseType(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDatabaseType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetDatabaseType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringDatabaseType(t *testing.T) {
	tests := []struct {
		name string
		db   DATABASE
	}{
		{"mongo", MONGODB},
		{"mysql", MYSQL},
		{"unknown", INVALID},
		{"invalid1", INVALID - 1},
		{"invalid2", MYSQL + 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.db.String() == "" {
				t.Errorf("String should not be empty")
			}
		})
	}
}
