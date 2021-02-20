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
	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/config"
	secretStoreConfig "github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"

	"github.com/stretchr/testify/mock"
)

type MockTokenProvider struct {
	mock.Mock
}

// Run see interface.go
func (p *MockTokenProvider) Run() error {
	// Boilerplate that returns whatever Mock.On().Returns() is configured for
	arguments := p.Called()
	return arguments.Error(0)
}

func (p *MockTokenProvider) SetConfiguration(secretConfig secretStoreConfig.SecretStoreInfo, tokenConfig config.TokenFileProviderInfo) {
	// Boilerplate that returns whatever Mock.On().Returns() is configured for
	p.Called(secretConfig, tokenConfig)
}
