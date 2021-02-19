//
// Copyright (c) 2021 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0'
//

package secretstore

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/mock"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
)

func TestInvalidProvider(t *testing.T) {
	serviceConfig := config.SecretStoreInfo{
		TokenProvider:     "does-not-exist",
		TokenProviderType: OneShotProvider,
	}
	mockExecRunner := mockExecRunner{}
	mockExecRunner.On("LookPath", serviceConfig.TokenProvider).
		Return("", errors.New("fake file does not exist"))
	cancel, err := testCommon(serviceConfig, &mockExecRunner)
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()
	require.Error(t, err)

	mockExecRunner.AssertExpectations(t)
}

func TestInvalidProviderType(t *testing.T) {
	serviceConfig := config.SecretStoreInfo{
		TokenProvider:     "success-executable",
		TokenProviderType: "simple",
	}
	mockExecRunner := mockExecRunner{}
	cancel, err := testCommon(serviceConfig, &mockExecRunner)
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()
	require.Error(t, err)
	mockExecRunner.AssertExpectations(t)
}

func TestNoConfig(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p := NewTokenProvider(ctx, logger.MockLogger{}, NewDefaultExecRunner())
	// don't call SetConfiguration()
	err := p.Launch()
	defer cancel()
	require.Error(t, err)
}

func TestSuccess(t *testing.T) {
	serviceConfig := config.SecretStoreInfo{
		TokenProvider:     "success-executable",
		TokenProviderType: OneShotProvider,
		TokenProviderArgs: []string{"arg1", "arg2"},
	}
	mockExecRunner := mockExecRunner{}
	mockCmd := mockCmd{}
	mockExecRunner.On("LookPath", serviceConfig.TokenProvider).
		Return(serviceConfig.TokenProvider, nil)
	mockExecRunner.On("CommandContext", mock.Anything,
		serviceConfig.TokenProvider, serviceConfig.TokenProviderArgs).
		Return(&mockCmd)
	mockCmd.On("Start").Return(nil)
	mockCmd.On("Wait").Return(nil)
	cancel, err := testCommon(serviceConfig, &mockExecRunner)
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()
	require.NoError(t, err)
	mockExecRunner.AssertExpectations(t)
	mockCmd.AssertExpectations(t)
}

func TestFailure(t *testing.T) {
	serviceConfig := config.SecretStoreInfo{
		TokenProvider:     "failure-executable",
		TokenProviderType: OneShotProvider,
		TokenProviderArgs: []string{"arg1", "arg2"},
	}
	mockExecRunner := mockExecRunner{}
	mockCmd := mockCmd{}
	mockExecRunner.On("LookPath", serviceConfig.TokenProvider).
		Return(serviceConfig.TokenProvider, nil)
	mockExecRunner.On("CommandContext", mock.Anything,
		serviceConfig.TokenProvider, serviceConfig.TokenProviderArgs).
		Return(&mockCmd)
	mockCmd.On("Start").Return(nil)
	mockCmd.On("Wait").Return(&exec.ExitError{ProcessState: &os.ProcessState{}})
	cancel, err := testCommon(serviceConfig, &mockExecRunner)
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()
	require.Error(t, err)
	mockExecRunner.AssertExpectations(t)
	mockCmd.AssertExpectations(t)
}

func testCommon(config config.SecretStoreInfo, mockExecRunner ExecRunner) (context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())
	p := NewTokenProvider(ctx, logger.MockLogger{}, mockExecRunner)
	if err := p.SetConfiguration(config); err != nil {
		return cancel, err
	}
	return cancel, p.Launch()
}
