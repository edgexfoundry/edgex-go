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

package gate

import (
	"context"
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
		{"Good: gate cmd empty option", []string{}, false},
		{"Bad: gate invalid option", []string{"--invalid=xxx"}, true},
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

type testConfig struct {
	testHost              string
	bootstrapperStartPort int
	registryReadyPort     int
	databaseReadyPort     int
	kongDBReadyPort       int
	readyToRunPort        int
}

func TestExecuteWithAllDependentsRun(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}
	testHost := "localhost"
	testConfig := &testConfig{
		testHost:              "localhost",
		bootstrapperStartPort: 28001,
		registryReadyPort:     28002,
		databaseReadyPort:     28003,
		kongDBReadyPort:       28004,
		readyToRunPort:        28009,
	}
	config := setupMockServiceConfigs(testConfig)

	type executeReturn struct {
		code int
		err  error
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	wg := &sync.WaitGroup{}
	gate, err := NewCommand(ctx, wg, lc, config, []string{})
	defer func() {
		cancelFunc()
		wg.Wait()
	}()
	require.NoError(t, err)
	require.NotNil(t, gate)
	require.Equal(t, "gate", gate.GetCommandName())
	execRet := make(chan executeReturn, 1)
	// in a separate go-routine since the listenTcp is a blocking call
	go func() {
		statusCode, err := gate.Execute()
		execRet <- executeReturn{code: statusCode, err: err}
	}()

	tcpSrvErr := make(chan error)
	// start up all other dependent mock services:
	go func() {
		tcpSrvErr <- tcp.NewTcpServer().StartListener(testConfig.registryReadyPort,
			lc, testHost)
	}()
	go func() {
		tcpSrvErr <- tcp.NewTcpServer().StartListener(testConfig.kongDBReadyPort,
			lc, testHost)
	}()
	go func() {
		tcpSrvErr <- tcp.NewTcpServer().StartListener(testConfig.databaseReadyPort,
			lc, testHost)
	}()

	select {
	case ret := <-execRet:
		require.NoError(t, ret.err)
		require.Equal(t, interfaces.StatusCodeExitNormal, ret.code)
	case err := <-tcpSrvErr:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		require.Fail(t, "security bootstrapper gate never returned")
	}
}

func setupMockServiceConfigs(testConf *testConfig) *config.ConfigurationStruct {
	conf := &config.ConfigurationStruct{}
	conf.StageGate = config.StageGateInfo{
		BootStrapper: config.BootStrapperInfo{
			Host:      testConf.testHost,
			StartPort: testConf.bootstrapperStartPort,
		},
		Registry: config.RegistryInfo{
			Host:      testConf.testHost,
			Port:      12001,
			ReadyPort: testConf.registryReadyPort,
		},
		Database: config.DatabaseInfo{
			Host:      testConf.testHost,
			Port:      12002,
			ReadyPort: testConf.databaseReadyPort,
		},
		KongDB: config.KongDBInfo{
			Host:      testConf.testHost,
			Port:      12003,
			ReadyPort: testConf.kongDBReadyPort,
		},
		Ready: config.ReadyInfo{
			ToRunPort: testConf.readyToRunPort,
		},
	}
	return conf
}
