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

	. "github.com/edgexfoundry/edgex-go/internal/security/authtokenloader"
	"github.com/stretchr/testify/assert"
)

func TestMockInterfaceType(t *testing.T) {
	// Typecast will fail if doesn't implement interface properly
	var iface AuthTokenLoader = &MockAuthTokenLoader{}
	assert.NotNil(t, iface)
}

func TestMockLoad(t *testing.T) {
	o := &MockAuthTokenLoader{}
	o.On("Load", "path").Return("abcd", nil)

	token, err := o.Load("path")

	o.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, "abcd", token)
}
