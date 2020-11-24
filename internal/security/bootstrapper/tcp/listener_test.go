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
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

func TestStartListener(t *testing.T) {
	lc := logger.MockLogger{}
	errs := make(chan error, 1)
	testPort := 12345
	srv := NewTcpServer()
	// in a separate goroutine since listener is blocking the main test thread
	go func() {
		errs <- srv.StartListener(testPort, lc, "")
	}()

	// in this test case we want to give some time for listener comes first
	time.Sleep(2 * time.Second)

	go func() {
		errs <- DialTcp("127.0.0.1", testPort, lc)
	}()

	select {
	case err := <-errs:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		require.Fail(t, "DialTcp never returned")
	}
}

func TestStartListenerAlreadyInUse(t *testing.T) {
	lc := logger.MockLogger{}
	errs := make(chan error, 1)
	testPort := 12347
	srv1 := NewTcpServer()
	go func() {
		errs <- srv1.StartListener(testPort, lc, "")
	}()

	// try to start another listener with the same port
	// this will cause an error
	srv2 := NewTcpServer()
	go func() {
		errs <- srv2.StartListener(testPort, lc, "")
	}()

	select {
	case err := <-errs:
		require.Error(t, err)
	case <-time.After(5 * time.Second):
		require.Fail(t, "none of Tcp listeners never started")
	}
}

func TestStartListenerWithDialFirst(t *testing.T) {
	lc := logger.MockLogger{}
	errs := make(chan error, 1)
	testPort := 12341
	srv := NewTcpServer()

	go func() {
		errs <- DialTcp("127.0.0.1", testPort, lc)
	}()

	// literally delay some time so that DialTcp always comes first
	time.Sleep(2 * time.Second)

	// in a separate goroutine since listener is blocking the main test thread
	go func() {
		errs <- srv.StartListener(testPort, lc, "")
	}()

	select {
	case err := <-errs:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		require.Fail(t, "DialTcp never returned")
	}
}
