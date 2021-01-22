/*******************************************************************************
 * Copyright 2021 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package help

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

// TestHelp tests functionality of help command
func TestHelp(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}

	// Act
	command, err := NewCommand(lc, config, []string{})
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
	config := &config.ConfigurationStruct{}

	// Act
	command, err := NewCommand(lc, config, []string{"-badarg"})

	// Assert
	require.Error(t, err)
	require.Nil(t, command)
}
