//
// Copyright (c) 2019 Intel Corporation
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
// SPDX-License-Identifier: Apache-2.0'
//
package fileprovider

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultTokenPolicy(t *testing.T) {
	// Act
	policies := makeDefaultTokenPolicy("service-name")

	// Assert
	bytes, err := json.Marshal(policies)
	assert.NoError(t, err)

	expected := `{"path":{"secret/edgex/service-name/*":{"capabilities":["create","update","delete","list","read"]}}}`
	actual := string(bytes)
	assert.Equal(t, expected, actual)
}

func TestDefaultTokenParameters(t *testing.T) {
	// Act
	parameters := makeDefaultTokenParameters("service-name")

	// Assert
	bytes, err := json.Marshal(parameters)
	assert.NoError(t, err)

	expected := `{"display_name":"service-name","no_parent":true,"policies":["edgex-service-service-name"]}`
	actual := string(bytes)
	assert.Equal(t, expected, actual)
}
