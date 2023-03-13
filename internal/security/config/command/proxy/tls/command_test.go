//
// Copyright (c) 2020-2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package tls

import (
	"bytes"
	"os"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/v3/pkg/token/fileioperformer/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTLSBagArguments tests command line errors
func TestTLSBagArguments(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}
	badArgTestcases := [][]string{
		{},                       // missing arg --in
		{"-badarg"},              // invalid arg
		{"--inCert", "somefile"}, // missing --inKey
		{"--inKey", "keyfile"},   // missing --inCert
		{"--inKey"},              // missing filename
		{"--inCert"},             // missing filename
		{"--targetFolder"},       // missing filename
		{"--certFilename"},       // missing filename
		{"--keyFilename"},        // missing filename
	}

	for _, args := range badArgTestcases {
		// Act
		command, err := NewCommand(lc, args)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, command)
	}
}

// TestTLSErrorFileNotFound tests the tls error regarding file not found issues
func TestTLSErrorFileNotFound(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}
	fileNotFoundTestcases := [][]string{
		{"--inCert", "missingcertificate", "--inKey", "missingprivatekey"},       // both files missing
		{"--inCert", "testdata/testCert.pem", "--inKey", "missingprivatekey"},    // key file missing
		{"--inCert", "missingcertificate", "--inKey", "testdata/testCert.prkey"}, // cert file missing
	}

	for _, args := range fileNotFoundTestcases {
		// Act
		command, err := NewCommand(lc, args)
		require.NoError(t, err)
		code, err := command.Execute()

		// Assert
		require.Error(t, err)
		require.Equal(t, interfaces.StatusCodeExitWithError, code)
	}
}

// TestInstallCertificate tests the happy path
func TestInstallCertificate(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}
	validCommandLines := [][]string{
		{"new.crt", "new.key", "", "", "/etc/ssl/nginx/nginx.crt", "/etc/ssl/nginx/nginx.key"},
		{"new.crt", "new.key", "--targetFolder", "/foofolder", "/foofolder/nginx.crt", "/foofolder/nginx.key"},
		{"new.crt", "new.key", "--certFilename", "foo.crt", "/etc/ssl/nginx/foo.crt", "/etc/ssl/nginx/nginx.key"},
		{"new.crt", "new.key", "--keyFilename", "foo.key", "/etc/ssl/nginx/nginx.crt", "/etc/ssl/nginx/foo.key"},
	}

	for _, args := range validCommandLines {
		// Arrange
		command, err := NewCommand(lc, []string{"--inCert", args[0], "--inKey", args[1], args[2], args[3]})
		require.NoError(t, err)
		mockOpener := &mocks.FileIoPerformer{}
		command.fileOpener = mockOpener

		mockOpener.On("OpenFileReader", args[0], os.O_RDONLY, os.FileMode(0400)).Return(bytes.NewBufferString("acert"), nil)
		mockOpener.On("OpenFileReader", args[1], os.O_RDONLY, os.FileMode(0400)).Return(bytes.NewBufferString("akey"), nil)
		mockOpener.On("OpenFileWriter", args[4], os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0644)).Return(&writeCloserBuffer{new(bytes.Buffer)}, nil)
		mockOpener.On("OpenFileWriter", args[5], os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0600)).Return(&writeCloserBuffer{new(bytes.Buffer)}, nil)

		// Act
		code, err := command.Execute()

		// Assert
		require.NoError(t, err)
		require.Equal(t, interfaces.StatusCodeExitNormal, code)
	}
}

//
// mocks
//

type writeCloserBuffer struct {
	*bytes.Buffer
}

func (wcb *writeCloserBuffer) Close() error {
	return nil
}

func (wcb *writeCloserBuffer) Chmod(_ os.FileMode) error {
	return nil
}

func (wcb *writeCloserBuffer) Chown(_ int, _ int) error {
	return nil
}
