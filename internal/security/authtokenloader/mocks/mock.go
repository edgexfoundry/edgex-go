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
	"github.com/stretchr/testify/mock"
)

type MockAuthTokenLoader struct {
	mock.Mock
}

func (m *MockAuthTokenLoader) Load(path string) (string, error) {
	// Boilerplate that returns whatever Mock.On().Returns() is configured for
	arguments := m.Called(path)
	return arguments.String(0), arguments.Error(1)
}
