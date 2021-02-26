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

package command

import (
	"context"
	"sync"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"

	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

func TestNewCommand(t *testing.T) {
	// Arrange
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{
		StageGate: config.StageGateInfo{
			WaitFor: config.WaitForInfo{
				Timeout:       "2s",
				RetryInterval: "1s",
			},
		},
	}

	tests := []struct {
		name            string
		cmdArgs         []string
		expectedCmdName string
		expectedErr     bool
	}{
		{"Good: gate command", []string{"gate"}, "gate", false},
		{"Good: pingPgDb command only", []string{"pingPgDb"}, "pingPgDb", false},
		{"Good: pingPgDb command with options", []string{"pingPgDb", "--host=kong-db", "--port=5432"}, "pingPgDb", false},
		{"Good: listenTcp command", []string{"listenTcp", "--port=55555"}, "listenTcp", false},
		{"Good: genPassword command", []string{"genPassword"}, "genPassword", false},
		{"Good: getHttpStatus command", []string{"getHttpStatus", "--url=http://localhost:55555"}, "getHttpStatus", false},
		{"Good: waitFor command", []string{"waitFor", "--uri=http://localhost:55555"}, "waitFor", false},
		{"Good: setupRegistryACL command", []string{"setupRegistryACL"}, "setupRegistryACL", false},
		{"Bad: unknown command", []string{"unknown"}, "", true},
		{"Bad: empty command", []string{}, "", true},
		{"Bad: listenTcp command missing required --port", []string{"listenTcp"}, "", true},
		{"Bad: getHttpStatus command missing required --url", []string{"getHttpStatus"}, "", true},
		{"Bad: waitFor command missing required --uri", []string{"waitFor"}, "", true},
	}

	for _, tt := range tests {
		test := tt //capture as local copy
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// Act
			command, err := NewCommand(ctx, wg, lc, config, test.cmdArgs)

			// Assert
			if test.expectedErr {
				require.Error(t, err)
				require.Nil(t, command)
			} else {
				require.NoError(t, err)
				require.NotNil(t, command)
				require.Equal(t, test.expectedCmdName, command.GetCommandName())
			}
		})
	}
}
