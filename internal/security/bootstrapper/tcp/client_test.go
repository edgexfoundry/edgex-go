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

package tcp

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

func TestDialTcpClient(t *testing.T) {
	lc := logger.MockLogger{}
	errs := make(chan error, 1)
	testListeningPort := 12333
	srv := NewTcpServer()
	go func() {
		errs <- srv.StartListener(testListeningPort, lc, "")
	}()

	time.Sleep(time.Second)

	tests := []struct {
		name        string
		host        string
		port        int
		expectError bool
	}{
		{"Good: Empty host input", "", testListeningPort, false},
		{"Bad: Port number = 0", "localhost", 0, true},
		{"Bad: Port number < 0", "localhost", -1, true},
		{"Bad: Both empty host and 0 port number", "", 0, true},
		{"Good: Dial the TCP server and port", "localhost", testListeningPort, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// dialTcp in goroutine so it won't block forever
			go func() {
				errs <- DialTcp(tt.host, tt.port, lc)
			}()

			select {
			case err := <-errs:
				if tt.expectError {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			case <-time.After(2 * time.Second):
				require.Fail(t, "DialTcp never returned")
			}
		})
	}
}

func TestDialTcpNoTCPServer(t *testing.T) {
	lc := logger.MockLogger{}
	errs := make(chan error, 1)
	testPort := 12349
	go func() {
		errs <- DialTcp("127.0.0.1", testPort, lc)
	}()

	select {
	case err := <-errs:
		require.NoError(t, err)

	// since the tcp server is never up, the dial will block forever,
	// so we set a timeout for the test to signal it is done
	case <-time.After(dialTimeoutDuration + time.Second):
		fmt.Println("Expected timed out due to no TCP server")
	}
}
