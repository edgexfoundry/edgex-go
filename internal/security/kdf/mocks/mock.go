//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockKeyDeriver struct {
	mock.Mock
}

func (m *MockKeyDeriver) DeriveKey(ikm []byte, keyLen uint, info string) ([]byte, error) {
	// Boilerplate that returns whatever Mock.On().Returns() is configured for
	arguments := m.Called(ikm, keyLen, info)
	return arguments.Get(0).([]byte), arguments.Error(1)
}
