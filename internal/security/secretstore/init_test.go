//
// Copyright (c) 2021 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package secretstore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/fileioperformer/mocks"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/stretchr/testify/assert"
)

const sampleJSON = `
{
	"keys": [
		"test-keys"
	],
	"keys_base64": [
		"test-keys-base64"
	],
	"root_token": "test-root-token"
}`

var expectedFolder = "/foo"
var expectedFile = "bar.baz"

func TestLoadInitResponse(t *testing.T) {
	// Arrange
	mockLogger := logger.MockLogger{}
	fileOpener := &mocks.FileIoPerformer{}
	stringReader := strings.NewReader(sampleJSON)
	fileOpener.On("OpenFileReader", filepath.Join(expectedFolder, expectedFile), os.O_RDONLY, os.FileMode(0400)).Return(stringReader, nil)
	secretConfig := config.SecretStoreInfo{
		TokenFolderPath: expectedFolder,
		TokenFile:       expectedFile,
	}
	initResponse := types.InitResponse{}

	// Act
	err := loadInitResponse(mockLogger, fileOpener, secretConfig, &initResponse)

	// Assert
	assert.NoError(t, err)
	fileOpener.AssertExpectations(t)
}

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
