//
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package secretstore

import (
	"os"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	. "github.com/edgexfoundry/go-mod-secrets/pkg/token/fileioperformer/mocks"

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

func TestLoadInitResponse(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	mockLogger := logger.MockLogger{}
	fileOpener := &MockFileIoPerformer{}
	stringReader := strings.NewReader(sampleJSON)
	fileOpener.On("OpenFileReader", "/foo/bar.baz", os.O_RDONLY, os.FileMode(0400)).Return(stringReader, nil)
	secretConfig := secretstoreclient.SecretServiceInfo{
		TokenFolderPath: "/foo",
		TokenFile:       "bar.baz",
	}
	initResponse := secretstoreclient.InitResponse{}

	// Act
	err := loadInitResponse(mockLogger, fileOpener, secretConfig, &initResponse)

	// Assert
	assert.NoError(err)
	fileOpener.AssertExpectations(t)
}

func TestSaveInitResponse(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	mockLogger := logger.MockLogger{}
	fileOpener := &MockFileIoPerformer{}
	fileOpener.On("OpenFileWriter", "/foo/bar.baz", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0600)).Return(&discardWriterCloser{}, nil)
	secretConfig := secretstoreclient.SecretServiceInfo{
		TokenFolderPath: "/foo",
		TokenFile:       "bar.baz",
	}
	initResponse := secretstoreclient.InitResponse{
		Keys:       []string{"test-key-1"},
		KeysBase64: []string{"dGVzdC1rZXktMQ=="},
	}

	// Act
	err := saveInitResponse(mockLogger, fileOpener, secretConfig, &initResponse)

	// Assert
	assert.NoError(err)
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
