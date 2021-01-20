//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package jwt

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJWTBadArguments tests command line errors
func TestJWTBadArguments(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}
	badArgTestcases := [][]string{
		{},                             // missing arg --algorithm
		{"-badarg"},                    // invalid arg
		{"--algorithm", "myalgorithm"}, // invalid --algorithm
		{"--algorithm", "RS256"},       // missing --private_key
		{"--algorithm", "RS256", "--private_key", "testdata/rsa.key"}, // missing --id
	}

	for _, args := range badArgTestcases {
		// Act
		command, err := NewCommand(lc, config, args)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, command)
	}
}

func generateWithArgs(t *testing.T, args []string) {
	// Arrange
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}

	// Act
	command, err := NewCommand(lc, config, args)
	require.NoError(t, err)
	code, err := command.Execute()

	// Assert
	require.NoError(t, err)
	require.Equal(t, interfaces.StatusCodeExitNormal, code)
}

// TestJWTGenerateRSA tests RSA JWT generation
func TestJWTGenerateRSA(t *testing.T) {
	generateWithArgs(t, []string{
		"--algorithm", "RS256",
		"--private_key", "testdata/rsa.key",
		"--id", "7f3ab74c-3bc2-4635-bc28-161f7f7ef246",
		"--expiration", "24h",
	})
}

// TestJWTGenerateECDSA tests ECDSA JWT generation
func TestJWTGenerateECDSA(t *testing.T) {
	generateWithArgs(t, []string{
		"--algorithm", "ES256",
		"--private_key", "testdata/ecdsa.key",
		"--id", "7f3ab74c-3bc2-4635-bc28-161f7f7ef246",
		"--expiration", "24h",
	})
}
