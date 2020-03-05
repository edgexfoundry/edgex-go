//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockPipedHexReader struct {
	mock.Mock
}

func (m *MockPipedHexReader) ReadHexBytesFromExe(executable string) ([]byte, error) {
	// Boilerplate that returns whatever Mock.On().Returns() is configured for
	arguments := m.Called(executable)
	return arguments.Get(0).([]byte), arguments.Error(1)
}
