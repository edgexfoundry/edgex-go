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

package gethttpstatus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/interfaces"

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
		{"Good: getHttpStatus required --url option", []string{"--url=http://localhost:32323"}, false},
		{"Bad: getHttpStatus invalid option", []string{"--invalid=http://localhost:123"}, true},
		{"Bad: getHttpStatus empty option", []string{""}, true},
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

	testSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer testSrv.Close()

	tests := []struct {
		name        string
		cmdArgs     []string
		expectedErr bool
	}{
		{"Good: getHttpStatus with existing server", []string{"--url=" + testSrv.URL}, false},
		{"Bad: getHttpStatus with non-existing server", []string{"--url=http://non-existing:1111"}, true},
		{"Bad: getHttpStatus with malformed URL", []string{"--url=_http!@xxxxxx:1111"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getHttpStatus, err := NewCommand(ctx, wg, lc, config, tt.cmdArgs)
			require.NoError(t, err)
			require.NotNil(t, getHttpStatus)
			require.Equal(t, "getHttpStatus", getHttpStatus.GetCommandName())

			statusCode, err := getHttpStatus.Execute()

			if tt.expectedErr {
				require.Error(t, err)
				require.Equal(t, interfaces.StatusCodeExitWithError, statusCode)
			} else {
				require.NoError(t, err)
				require.Equal(t, interfaces.StatusCodeExitNormal, statusCode)
			}
		})
	}
}
