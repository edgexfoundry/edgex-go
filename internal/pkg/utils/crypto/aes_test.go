//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package crypto

import (
	"testing"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestNewAESCryptorWithSecretProvider_NilProvider(t *testing.T) {
	cryptor, err := NewAESCryptorWithSecretProvider(nil)
	require.Error(t, err)
	assert.Nil(t, cryptor)
}

func TestNewAESCryptorWithSecretProvider_WithProvider(t *testing.T) {
	mockProvider := &mocks.SecretProviderExt{}
	mockProvider.On("GetSecret", aesSecretName).Return(nil, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "secret not found", nil))
	mockProvider.On("StoreSecret", aesSecretName, mock.AnythingOfType("map[string]string")).Return(nil)

	cryptor, err := NewAESCryptorWithSecretProvider(mockProvider)
	require.NoError(t, err)
	assert.NotNil(t, cryptor)

	// Verify that GetSecret and StoreSecret was called
	mockProvider.AssertExpectations(t)

	// Test encryption and decryption
	plaintext := "test message"
	encrypted, err := cryptor.Encrypt(plaintext)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := cryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, string(decrypted))
}

func TestNewAESCryptorWithSecretProvider_WithExistingKey(t *testing.T) {
	mockProvider := &mocks.SecretProviderExt{}
	existingKey := aesKey

	mockProvider.On("GetSecret", aesSecretName).Return(map[string]string{
		aesKeyName: existingKey,
	}, nil)

	cryptor, err := NewAESCryptorWithSecretProvider(mockProvider)
	require.NoError(t, err)
	assert.NotNil(t, cryptor)

	// Verify that GetSecret was called but StoreSecret was not called
	mockProvider.AssertExpectations(t)
	mockProvider.AssertNotCalled(t, "StoreSecret")

	// Test encryption and decryption
	plaintext := "test message"
	encrypted, err := cryptor.Encrypt(plaintext)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := cryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, string(decrypted))
}
