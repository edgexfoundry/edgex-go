//
// Copyright (c) 2020-2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package help

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"

	"github.com/stretchr/testify/require"
)

// TestHelp tests functionality of help command
func TestHelp(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}

	// Act
	command, err := NewCommand(lc, []string{})
	require.NoError(t, err)

	code, err := command.Execute()

	// Assert
	require.NoError(t, err)
	require.Equal(t, interfaces.StatusCodeExitNormal, code)
}

// TestHelpBadArg tests unknown arg handler
func TestHelpBadArg(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}

	// Act
	command, err := NewCommand(lc, []string{"-badarg"})

	// Assert
	require.Error(t, err)
	require.Nil(t, command)
}
