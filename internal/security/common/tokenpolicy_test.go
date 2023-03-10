// Copyright (c) 2019-2023 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultTokenPolicy(t *testing.T) {
	// Act
	policies := MakeDefaultTokenPolicy("service-name")

	// Assert
	bytes, err := json.Marshal(policies)
	require.NoError(t, err)
	require.NotEmpty(t, bytes)

	expected := map[string]interface{}{
		"path": map[string]interface{}{
			"secret/edgex/service-name/*": map[string]interface{}{
				"capabilities": []string{"create", "update", "delete", "list", "read"},
			},
			"consul/creds/service-name": map[string]interface{}{
				"capabilities": []string{"read"},
			},
			"identity/oidc/token/service-name": map[string]interface{}{
				"capabilities": []string{"read"},
			},
			"identity/oidc/introspect": map[string]interface{}{
				"capabilities": []string{"create", "update"},
			},
		},
	}

	require.Equal(t, expected, policies)
}
