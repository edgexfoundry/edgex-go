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

package device

import (
	"context"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

func TestNewRequester(t *testing.T) {
	var invalidType RequestType
	invalidType = 3

	tests := []struct {
		name        string
		rt          RequestType
		expectError bool
	}{
		{"http_ok", Http, false},
		{"mock_ok", Mock, false},
		{"invalid_!ok", invalidType, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRequester(tt.rt, logger.MockLogger{}, context.Background())
			if err != nil {
				if !tt.expectError {
					t.Errorf("%s unexpected error: %v", tt.name, err)
				}
			} else {
				if tt.expectError {
					t.Errorf("did not receive expected error: %s", tt.name)
				}
			}
		})
	}
}
