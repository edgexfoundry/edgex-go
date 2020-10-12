//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package mocks

import (
	"testing"

	. "github.com/edgexfoundry/edgex-go/internal/security/kdf"
	"github.com/stretchr/testify/assert"
)

func TestMockInterfaceType(t *testing.T) {
	// Typecast will fail if doesn't implement interface properly
	var iface KeyDeriver = &MockKeyDeriver{}
	assert.NotNil(t, iface)
}

func TestDeriveKey(t *testing.T) {
	mockClient := &MockKeyDeriver{}
	mockClient.On("DeriveKey", make([]byte, 32), uint(32), "info").Return(make([]byte, 1), nil)

	ikm, err := mockClient.DeriveKey(make([]byte, 32), 32, "info")
	assert.Nil(t, err)
	assert.Equal(t, make([]byte, 1), ikm)
	mockClient.AssertExpectations(t)
}
