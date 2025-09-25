//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package crypto

import (
	"fmt"
	"testing"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var defaultKey = []byte("qW8rMz2bXe4uVk1nL9aDpTYoJhCg7RsF")

func TestAESCryptor_Encryption(t *testing.T) {
	testData := "test data"
	aesCryptor := NewAESCryptor(defaultKey)

	encrypted, err := aesCryptor.Encrypt(testData)
	require.NoError(t, err)

	decrypted, err := aesCryptor.Decrypt(encrypted)
	require.NoError(t, err)
	require.Equal(t, testData, string(decrypted))
}

func TestNewAESCryptorWithSecretProvider_NilProvider(t *testing.T) {
	cryptor, err := NewAESCryptorWithSecretProvider(nil, defaultKey)
	require.Error(t, err)
	assert.Nil(t, cryptor)
	assert.Contains(t, err.Error(), "secret provider is nil")
}

func TestNewAESCryptorWithSecretProvider_WithProvider(t *testing.T) {
	mockProvider := &mocks.SecretProvider{}
	mockProvider.On("GetSecret", aesSecretName).Return(nil, pkg.NewErrSecretNameNotFound(aesSecretName))
	mockProvider.On("StoreSecret", aesSecretName, mock.AnythingOfType("map[string]string")).Return(nil)

	cryptor, err := NewAESCryptorWithSecretProvider(mockProvider, defaultKey)
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
	mockProvider := &mocks.SecretProvider{}

	mockProvider.On("GetSecret", aesSecretName).Return(map[string]string{
		aesKeyName: string(defaultKey),
	}, nil)

	cryptor, err := NewAESCryptorWithSecretProvider(mockProvider, defaultKey)
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

func TestNewAESCryptorWithSecretProvider_EmptyKeyInSecretStore(t *testing.T) {
	mockProvider := &mocks.SecretProvider{}

	mockProvider.On("GetSecret", aesSecretName).Return(map[string]string{
		aesKeyName: "", // Empty key
	}, nil)
	mockProvider.On("StoreSecret", aesSecretName, mock.AnythingOfType("map[string]string")).Return(nil)

	cryptor, err := NewAESCryptorWithSecretProvider(mockProvider, defaultKey)
	require.NoError(t, err)
	assert.NotNil(t, cryptor)

	// Verify that GetSecret was called and StoreSecret was called
	mockProvider.AssertExpectations(t)
}

func TestNewAESCryptorWithSecretProvider_MissingKeyInSecretStore(t *testing.T) {
	mockProvider := &mocks.SecretProvider{}

	mockProvider.On("GetSecret", aesSecretName).Return(map[string]string{
		"other_key": "some_value",
	}, nil)
	mockProvider.On("StoreSecret", aesSecretName, mock.AnythingOfType("map[string]string")).Return(nil)

	cryptor, err := NewAESCryptorWithSecretProvider(mockProvider, defaultKey)
	require.NoError(t, err)
	assert.NotNil(t, cryptor)

	// Verify that GetSecret was called and StoreSecret was called
	mockProvider.AssertExpectations(t)
}

func TestNewAESCryptorWithSecretProvider_InvalidBase64Key(t *testing.T) {
	mockProvider := &mocks.SecretProvider{}
	invalidBase64Key := "invalid-base64-key!"

	mockProvider.On("GetSecret", aesSecretName).Return(map[string]string{
		aesKeyName: invalidBase64Key,
	}, nil)

	cryptor, err := NewAESCryptorWithSecretProvider(mockProvider, defaultKey)
	require.Error(t, err)
	assert.Nil(t, cryptor)
	assert.Contains(t, err.Error(), "invalid AES key format in secret store")

	// Verify that GetSecret was called but StoreSecret was not called
	mockProvider.AssertExpectations(t)
	mockProvider.AssertNotCalled(t, "StoreSecret")
}

func TestNewAESCryptorWithSecretProvider_GetSecretError(t *testing.T) {
	mockProvider := &mocks.SecretProvider{}
	expectedErrorString := "failed to get AES key from secret store"

	mockProvider.On("GetSecret", aesSecretName).Return(nil, fmt.Errorf("server error"))

	cryptor, err := NewAESCryptorWithSecretProvider(mockProvider, defaultKey)
	require.Error(t, err)
	assert.Nil(t, cryptor)
	assert.Contains(t, err.Error(), expectedErrorString)

	// Verify that StoreSecret was not called
	mockProvider.AssertNotCalled(t, "StoreSecret")
}

func TestNewAESCryptorWithSecretProvider_StoreSecretError(t *testing.T) {
	mockProvider := &mocks.SecretProvider{}
	expectedErrorString := "failed to store AES key in secret store"

	mockProvider.On("GetSecret", aesSecretName).Return(nil, pkg.NewErrSecretNameNotFound(aesSecretName))
	mockProvider.On("StoreSecret", aesSecretName, mock.AnythingOfType("map[string]string")).Return(
		fmt.Errorf("server error"))

	cryptor, err := NewAESCryptorWithSecretProvider(mockProvider, defaultKey)
	require.Error(t, err)
	assert.Nil(t, cryptor)
	assert.Contains(t, err.Error(), expectedErrorString)
}
