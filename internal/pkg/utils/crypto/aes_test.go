//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAESCryptor_Encryption(t *testing.T) {
	testData := "test data"
	aesCryptor := NewAESCryptor()

	encrypted, err := aesCryptor.Encrypt(testData)
	require.NoError(t, err)

	decrypted, err := aesCryptor.Decrypt(encrypted)
	require.NoError(t, err)
	require.Equal(t, testData, string(decrypted))
}
