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
// SPDX-License-Identifier: Apache-2.0
//

package mocks

import (
	"testing"

	. "github.com/edgexfoundry/edgex-go/internal/security/fileprovider"
	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/config"

	secretStoreConfig "github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"

	"github.com/stretchr/testify/assert"
)

func TestMockInterfaceType(t *testing.T) {
	// Typecast will fail if doesn't implement interface properly
	var provider TokenProvider = &MockTokenProvider{}
	assert.NotNil(t, provider)
}

func TestMockRun(t *testing.T) {
	p := &MockTokenProvider{}
	p.On("Run").Return(nil)

	err := p.Run()

	assert.Nil(t, err)
	p.AssertExpectations(t)
}

func TestMockSetConfiguration(t *testing.T) {
	p := &MockTokenProvider{}
	p.On("SetConfiguration", secretStoreConfig.SecretStoreInfo{}, config.TokenFileProviderInfo{})

	p.SetConfiguration(secretStoreConfig.SecretStoreInfo{}, config.TokenFileProviderInfo{})

	p.AssertExpectations(t)
}
