//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package mocks

import (
	"testing"

	. "github.com/edgexfoundry/edgex-go/internal/security/pipedhexreader"
	"github.com/stretchr/testify/assert"
)

func TestMockInterfaceType(t *testing.T) {
	// Typecast will fail if doesn't implement interface properly
	var iface PipedHexReader = &MockPipedHexReader{}
	assert.NotNil(t, iface)
}

func TestReadHexBytesFromExe(t *testing.T) {
	mockClient := &MockPipedHexReader{}
	mockClient.On("ReadHexBytesFromExe", "/bin/somexe").Return(make([]byte, 1), nil)

	ikm, err := mockClient.ReadHexBytesFromExe("/bin/somexe")
	assert.Nil(t, err)
	assert.Equal(t, make([]byte, 1), ikm)
	mockClient.AssertExpectations(t)
}
