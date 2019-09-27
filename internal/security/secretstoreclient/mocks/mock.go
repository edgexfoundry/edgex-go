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
	"io"

	. "github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"
	"github.com/stretchr/testify/mock"
)

type MockSecretStoreClient struct {
	mock.Mock
}

func (m *MockSecretStoreClient) HealthCheck() (statusCode int, err error) {
	// Boilerplate that returns whatever Mock.On().Returns() is configured for
	arguments := m.Called()
	return arguments.Int(0), arguments.Error(1)
}

func (m *MockSecretStoreClient) Init(config SecretServiceInfo, vmkWriter io.Writer) (statusCode int, err error) {
	// Boilerplate that returns whatever Mock.On().Returns() is configured for
	arguments := m.Called(config, vmkWriter)
	return arguments.Int(0), arguments.Error(1)
}

func (m *MockSecretStoreClient) Unseal(config SecretServiceInfo, vmkReader io.Reader) (statusCode int, err error) {
	// Boilerplate that returns whatever Mock.On().Returns() is configured for
	arguments := m.Called(config, vmkReader)
	return arguments.Int(0), arguments.Error(1)
}

func (m *MockSecretStoreClient) InstallPolicy(token string, policyName string, policyDocument string) (statusCode int, err error) {
	// Boilerplate that returns whatever Mock.On().Returns() is configured for
	arguments := m.Called(token, policyName, policyDocument)
	return arguments.Int(0), arguments.Error(1)
}

func (m *MockSecretStoreClient) CreateToken(token string, parameters map[string]interface{}, response interface{}) (statusCode int, err error) {
	// Boilerplate that returns whatever Mock.On().Returns() is configured for
	arguments := m.Called(token, parameters, response)
	return arguments.Int(0), arguments.Error(1)
}
