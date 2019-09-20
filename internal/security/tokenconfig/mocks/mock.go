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
	. "github.com/edgexfoundry/edgex-go/internal/security/tokenconfig"
	"github.com/stretchr/testify/mock"
)

type MockTokenConfigParser struct {
	mock.Mock
}

func (m *MockTokenConfigParser) Load(path string) error {
	// Boilerplate that returns whatever Mock.On().Returns() is configured for
	arguments := m.Called(path)
	return arguments.Error(0)
}

func (m *MockTokenConfigParser) ServiceKeys() []string {
	// Boilerplate that returns whatever Mock.On().Returns() is configured for
	arguments := m.Called()
	return arguments.Get(0).([]string)
}

func (m *MockTokenConfigParser) GetServiceConfig(service string) ServiceKey {
	// Boilerplate that returns whatever Mock.On().Returns() is configured for
	arguments := m.Called(service)
	return arguments.Get(0).(ServiceKey)
}
