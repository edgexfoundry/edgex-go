//
// Copyright (c) 2021-2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package secretstore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/token/fileioperformer/mocks"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedFolder = "/foo"
var expectedFile = "bar.baz"

func TestSaveInitResponse(t *testing.T) {

	// Arrange
	mockLogger := logger.MockLogger{}
	fileOpener := &mocks.FileIoPerformer{}
	fileOpener.On("OpenFileWriter", filepath.Join(expectedFolder, expectedFile), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0600)).Return(&discardWriterCloser{}, nil)
	secretConfig := config.SecretStoreInfo{
		TokenFolderPath: expectedFolder,
		TokenFile:       expectedFile,
	}
	initResponse := types.InitResponse{
		Keys:       []string{"test-key-1"},
		KeysBase64: []string{"dGVzdC1rZXktMQ=="},
	}

	// Act
	err := saveInitResponse(mockLogger, fileOpener, secretConfig, &initResponse)

	// Assert
	assert.NoError(t, err)
	fileOpener.AssertExpectations(t)
}

func TestGetKnownSecretsToAdd(t *testing.T) {
	defer func() {
		_ = os.Setenv(addKnownSecretsEnv, "")
	}()

	expectedEmpty := map[string][]string{}
	expectedOneService := map[string][]string{
		"redisdb": {"service-1"},
	}
	expectedMultiServices := map[string][]string{
		"redisdb": {"service-1", "service-2", "service-3"},
	}

	tests := []struct {
		name                  string
		envValue              string
		expected              map[string][]string
		expectedErrorContains string
	}{
		{"valid empty", "", expectedEmpty, ""},
		{"valid one service", "redisdb[service-1]", expectedOneService, ""},
		{"valid multi services", "redisdb[service-1; service-2; service-3]", expectedMultiServices, ""},
		{"valid secret listed twice", "redisdb[service-1], redisdb[service-2; service-3]", expectedMultiServices, ""},
		{"invalid no services", "redisdb[]", nil, "list for 'redisdb' is empty"},
		{"invalid unknown secret", "messagebus[service-1; service-2; service-3]", nil, "'messagebus' is not a known secret"},
		{"invalid known & unknown secret", "redisdb[service-1], messagebus[service-1; service-2; service-3]", nil, "'messagebus' is not a known secret"},
		{"invalid service list, missing ]", "redisdb[service-1; service-2; service-3", nil, "Service list for 'redisdb' missing closing ']'"},
		{"invalid service list, missing [", "redisdb:service-1; service-2; service-3]", nil, "is invalid. Missing or too many '['"},
		{"invalid service name", "redisdb[service-%1]", nil, "Service name 'service-%1' has invalid characters"},
	}

	b := NewBootstrap(false, 10)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = os.Setenv(addKnownSecretsEnv, test.envValue)
			actual, err := b.getKnownSecretsToAdd()
			if test.expected == nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErrorContains)
				return
			}

			assert.Equal(t, test.expected, actual)
		})
	}
}

//
// mocks
//

type discardWriterCloser struct{}

func (wcb *discardWriterCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (wcb *discardWriterCloser) Close() error {
	return nil
}
