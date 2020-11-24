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

package listen

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/tcp"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

func TestNewCommand(t *testing.T) {
	// Arrange
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}

	tests := []struct {
		name        string
		cmdArgs     []string
		expectedErr bool
	}{
		{"Good: listenTcp required --port option", []string{"--port=32323"}, false},
		{"Good: listenTcp both options", []string{"--host=test", "--port=32323"}, false},
		{"Bad: listenTcp invalid option", []string{"--invalid=xxxxx"}, true},
		{"Bad: listenTcp empty option", []string{""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, err := NewCommand(ctx, wg, lc, config, tt.cmdArgs)
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, command)
			}
		})
	}
}

func TestExecute(t *testing.T) {
	// Arrange
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}
	testListenPort1 := 65111
	testListenPort2 := 65112

	tests := []struct {
		name        string
		cmdArgs     []string
		testPort    int
		expectedErr bool
	}{
		{"Good: listenTcp with unbound port",
			[]string{"--port=" + strconv.Itoa(testListenPort1)}, testListenPort1, false},
		{"Good: listenTcp with unbound port specific host",
			[]string{"--host=localhost", "--port=" + strconv.Itoa(testListenPort2)}, testListenPort2, false},
		{"Bad: listenTcp with already bound port",
			[]string{"--host=localhost", "--port=" + strconv.Itoa(testListenPort2)}, testListenPort2, true},
	}

	type executeReturn struct {
		code int
		err  error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listenTcp, err := NewCommand(ctx, wg, lc, config, tt.cmdArgs)
			require.NoError(t, err)
			require.NotNil(t, listenTcp)
			require.Equal(t, "listenTcp", listenTcp.GetCommandName())

			execRet := make(chan executeReturn, 1)
			// in a separate go-routine since the listenTcp is a blocking call
			go func() {
				statusCode, err := listenTcp.Execute()
				execRet <- executeReturn{code: statusCode, err: err}
			}()

			dialErr := make(chan error)
			// dial to tcp server to check the server being listening
			go func() {
				dialErr <- tcp.DialTcp("", tt.testPort, lc)
			}()

			select {
			case testErr := <-dialErr:
				require.NoError(t, testErr)
			case <-time.After(3 * time.Second):
				require.Fail(t, "DialTcp never returned")
			}

			// test to wait for some time to check the running tcp server is not errorred out
			// since receiving execRet channel will block forever if no error occurs
			select {
			case ret := <-execRet:
				if tt.expectedErr {
					require.Error(t, ret.err)
					require.Equal(t, interfaces.StatusCodeExitWithError, ret.code)
				} else {
					require.NoError(t, ret.err)
					require.Equal(t, interfaces.StatusCodeExitNormal, ret.code)
				}
			case <-time.After(5 * time.Second):
				t.Logf("tcp server %s listening ok", tt.cmdArgs)
			}
		})
	}
}
