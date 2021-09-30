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

package waitfor

import (
	"context"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	testDefaultTimeout := "10s"
	testDefaultRetryInterval := "1s"

	tests := []struct {
		name          string
		cmdArgs       []string
		timeout       string
		retryInterval string
		expectedErr   bool
	}{
		{"Good: waitFor required at least one --uri option", []string{"--uri=http://localhost:11120"},
			testDefaultTimeout, testDefaultRetryInterval, false},
		{"Good: waitFor multiple --uri options", []string{"--uri=http://localhost:11120", "--uri=file:///testfile"},
			testDefaultTimeout, testDefaultRetryInterval, false},
		{"Good: waitFor --uri with --timeout options", []string{"--uri=http://:11120", "--timeout=1s"},
			testDefaultTimeout, testDefaultRetryInterval, false},
		{"Good: waitFor multiple --uri with --timeout options",
			[]string{"--uri=http://:11120", "--uri=file:///testfile", "--timeout=1s"}, testDefaultTimeout,
			testDefaultRetryInterval, false},
		{"Good: waitFor --uri with --retryInterval options", []string{"--uri=http://:11120", "--retryInterval=5s"},
			testDefaultTimeout, testDefaultRetryInterval, false},
		{"Good: waitFor multiple --uri with --retryInterval options",
			[]string{"--uri=http://:11120", "--uri=file:///testfile", "--retryInterval=5s"}, testDefaultTimeout,
			testDefaultRetryInterval, false},
		{"Good: waitFor --uri --timeout --retryInterval options",
			[]string{"--uri=http://:11120", "--timeout=1s", "--retryInterval=5s"}, testDefaultTimeout,
			testDefaultRetryInterval, false},
		{"Bad: waitFor invalid option", []string{"--invalid=http://localhost:123"}, testDefaultTimeout,
			testDefaultRetryInterval, true},
		{"Bad: waitFor empty option", []string{""}, testDefaultTimeout, testDefaultRetryInterval, true},
		{"Bad: waitFor interval option parse error", []string{"--uri=http://:11120", "--timeout=100"},
			testDefaultTimeout, testDefaultRetryInterval, true},
		{"Bad: waitFor bad syntax timeout config", []string{"--uri=http://localhost:11120"}, "10",
			testDefaultRetryInterval, true},
		{"Bad: waitFor bad syntax retryInterval config", []string{"--uri=http://localhost:11120"}, testDefaultTimeout,
			"1", true},
		{"Bad: waitFor negative value timeout config", []string{"--uri=http://localhost:11120"}, "-10s",
			testDefaultRetryInterval, true},
		{"Bad: waitFor negative value retryInterval config", []string{"--uri=http://localhost:11120"},
			testDefaultTimeout, "-1m", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := getTestConfig(tt.timeout, tt.retryInterval)
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
	config := getTestConfig("7s", "1s")

	defaultWaitTimeout, err := time.ParseDuration(config.StageGate.WaitFor.Timeout)
	if err != nil {
		require.NoError(t, err)
	}

	testHost := "localhost"
	testPort := 11122
	testSrv := net.JoinHostPort(testHost, strconv.Itoa(testPort))

	testFile := "command_test.go"
	path, err := os.Getwd()

	if err != nil {
		require.NoError(t, err)
	}

	waitFile := filepath.Join(path, testFile)

	testHttpSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer testHttpSrv.Close()

	tests := []struct {
		name        string
		cmdArgs     []string
		testTimeout time.Duration
		expectedErr bool
	}{
		{"Good: waitFor with existing tcp server", []string{"--uri=tcp://" + testSrv}, defaultWaitTimeout, false},
		{"Good: waitFor with existing tcp server with timeout", []string{"--uri=tcp://" + testSrv, "--timeout=5s"},
			time.Duration(5 * time.Second), false},
		{"Good: waitFor with existing tcp server with retryInterval", []string{"--uri=tcp://" + testSrv, "--retryInterval=2s"},
			time.Duration(6 * time.Second), false},
		{"Good: waitFor with existing testFile",
			[]string{"--uri=file://" + waitFile}, defaultWaitTimeout, false},
		{"Good: waitFor with existing tcp server and testFile",
			[]string{"--uri=tcp://" + testSrv, "--uri=file://" + waitFile}, defaultWaitTimeout, false},
		{"Good: waitFor with existing tcp server and testFile and timeout",
			[]string{"--uri=tcp://" + testSrv, "--uri=file://" + waitFile, "--timeout=5s"}, defaultWaitTimeout, false},
		{"Good: waitFor with existing tcp server and testFile and retryInterval",
			[]string{"--uri=tcp://" + testSrv, "--uri=file://" + waitFile, "--retryInterval=2s"}, defaultWaitTimeout, false},
		{"Good: waitFor with existing http server", []string{"--uri=" + testHttpSrv.URL}, defaultWaitTimeout, false},
		{"Bad: waitFor with malformed URL", []string{"--uri=_http!@xxxxxx:1111"}, time.Duration(2 * time.Second), true},
		{"Bad: waitFor with no tcp server response", []string{"--uri=tcp://non-existing:1111", "--timeout=3s"},
			time.Duration(3 * time.Second), true},
		{"Bad: waitFor with no http server response", []string{"--uri=http://non-existing:1111", "--timeout=3s"},
			time.Duration(3 * time.Second), true},
		{"Bad: waitFor with no file", []string{"--uri=file://non-existing-file", "--timeout=3s"},
			time.Duration(3 * time.Second), true},
		{"Bad: waitFor with unsupported protocol", []string{"--uri=chrome://settings", "--timeout=3s"},
			time.Duration(3 * time.Second), true},
	}

	type execRet struct {
		statusCode int
		execErr    error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			waitFor, err := NewCommand(ctx, wg, lc, config, tt.cmdArgs)
			require.NoError(t, err)
			require.NotNil(t, waitFor)
			require.Equal(t, "waitFor", waitFor.GetCommandName())

			ret := make(chan execRet, 1)
			go func() {
				code, err := waitFor.Execute()
				ret <- execRet{statusCode: code, execErr: err}
			}()

			// literally make random delay between 0 to 3 seconds before running the tcp server
			time.Sleep(time.Duration(rand.Intn(3)) * time.Second) // nolint:gosec

			tcpSrvErr := make(chan error, 1)
			go func() {
				tcpSrvErr <- tcp.NewTcpServer().StartListener(testPort,
					lc, testHost)
			}()

			select {
			case res := <-ret:
				if tt.expectedErr {
					require.Error(t, res.execErr)
					require.Equal(t, interfaces.StatusCodeExitWithError, res.statusCode)
				} else {
					require.NoError(t, res.execErr)
					require.Equal(t, interfaces.StatusCodeExitNormal, res.statusCode)
				}
			case <-time.After(tt.testTimeout + time.Second):
				require.Fail(t, "waitFor failed to get response after test timeout")
			}
		})
	}
}

func getTestConfig(timeout, retryInterval string) *config.ConfigurationStruct {
	return &config.ConfigurationStruct{
		StageGate: config.StageGateInfo{
			WaitFor: config.WaitForInfo{
				Timeout:       timeout,
				RetryInterval: retryInterval,
			},
		},
	}
}
