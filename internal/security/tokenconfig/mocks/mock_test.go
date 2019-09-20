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

package mocks

import (
	"testing"

	. "github.com/edgexfoundry/edgex-go/internal/security/tokenconfig"
	"github.com/stretchr/testify/assert"
)

func TestMockInterfaceType(t *testing.T) {
	// Typecast will fail if doesn't implement interface properly
	var iface TokenConfigParser = &MockTokenConfigParser{}
	assert.NotNil(t, iface)
}

func TestMockLoad(t *testing.T) {
	parser := &MockTokenConfigParser{}
	parser.On("Load", "path").Return(nil)

	err := parser.Load("path")

	assert.Nil(t, err)
	parser.AssertExpectations(t)
}

func TestMockServiceKeys(t *testing.T) {
	parser := &MockTokenConfigParser{}
	parser.On("ServiceKeys").Return([]string{"key1"})

	keys := parser.ServiceKeys()

	assert.Equal(t, "key1", keys[0])
	parser.AssertExpectations(t)
}

func TestMockGetServiceConfig(t *testing.T) {
	parser := &MockTokenConfigParser{}
	parser.On("GetServiceConfig", "key1").Return(ServiceKey{UseDefaults: true})

	sk := parser.GetServiceConfig("key1")

	assert.True(t, sk.UseDefaults)
	parser.AssertExpectations(t)
}
