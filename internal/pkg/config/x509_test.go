/*******************************************************************************
 * Copyright 2019 Dell Inc.
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
 *******************************************************************************/

package config

import (
	"testing"
)

func TestLoadX509Config(t *testing.T) {
	tests := []struct {
		name        string
		file        string
		expectError bool
	}{
		{"509LoadOK", "./res/pkisetup-test.json", false},
		{"509NonexistentFile", "./res/pkisetup-nope.json", true},
		{"509InvalidJSON", "./res/pkisetup-fail.json", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewX509Config(tt.file)
			if err != nil && !tt.expectError {
				t.Errorf("unexpected error %v", err)
				return
			}
			if err == nil && tt.expectError {
				t.Error("expected error, none returned")
				return
			}
		})
	}
}
