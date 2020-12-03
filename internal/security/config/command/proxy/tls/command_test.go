//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package tls

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTLSBagArguments tests command line errors
func TestTLSBagArguments(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}
	badArgTestcases := [][]string{
		{},                       // missing arg --in
		{"-badarg"},              // invalid arg
		{"--incert", "somefile"}, // missing --inkey
	}

	for _, args := range badArgTestcases {
		// Act
		command, err := NewCommand(lc, config, args)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, command)
	}
}

// TestTLSErrorStub tests the tls error stub
func TestTLSErrorStub(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}
	args := []string{
		"--incert", "somecertificate",
		"--inkey", "someprivatekey",
	}

	// Act
	command, err := NewCommand(lc, config, args)
	require.NoError(t, err)
	code, err := command.Execute()

	// Assert
	require.Error(t, err)
	require.Equal(t, interfaces.StatusCodeExitWithError, code)
}
