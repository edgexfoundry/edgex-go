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

package main

import (
	"errors"
	"testing"

	. "github.com/edgexfoundry/edgex-go/internal/security/fileprovider/mocks"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNoOption(t *testing.T) {
	assert := assert.New(t)

	mockLogger := logger.MockLogger{}
	mockTokenProvider := &MockTokenProvider{}
	mockTokenProvider.On("Run").Return(nil)
	mockTokenProvider.On("SetConfiguration", mock.Anything, mock.Anything).Once()

	fileProvider = mockTokenProvider // fileProvider is global in main.go
	status := runTokenProvider(mockLogger)

	assert.False(status)
	mockTokenProvider.AssertExpectations(t)
}

func TestHandleError(t *testing.T) {
	assert := assert.New(t)

	mockLogger := logger.MockLogger{}
	mockTokenProvider := &MockTokenProvider{}
	mockTokenProvider.On("Run").Return(errors.New("an error occurred"))
	mockTokenProvider.On("SetConfiguration", mock.Anything, mock.Anything).Once()

	fileProvider = mockTokenProvider // fileProvider is global in main.go
	status := runTokenProvider(mockLogger)

	assert.False(status)
	mockTokenProvider.AssertExpectations(t)
}
