//
// Copyright (c) 2019 Intel Corporation
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

	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInvalidProvider(t *testing.T) {
	config := secretstoreclient.SecretServiceInfo{
		TokenProvider:     "does-not-exist",
		TokenProviderType: OneShotProvider,
	}
	mockExecRunner := mockExecRunner{}
	mockExecRunner.On("LookPath", config.TokenProvider).
		Return("", errors.New("fake file does not exist"))
	cancel, err := testCommon(config, &mockExecRunner)
	defer cancel()
	assert.NotNil(t, err)
	mockExecRunner.AssertExpectations(t)
}

func TestInvalidProviderType(t *testing.T) {
	config := secretstoreclient.SecretServiceInfo{
		TokenProvider:     "success-executable",
		TokenProviderType: "simple",
	}
	mockExecRunner := mockExecRunner{}
	cancel, err := testCommon(config, &mockExecRunner)
	defer cancel()
	assert.Error(t, err)
	mockExecRunner.AssertExpectations(t)
}

func TestNoConfig(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p := NewTokenProvider(ctx, logger.MockLogger{}, NewDefaultExecRunner())
	// don't call SetConfiguration()
	err := p.Launch()
	defer cancel()
	assert.Error(t, err)
}

func TestSuccess(t *testing.T) {
	config := secretstoreclient.SecretServiceInfo{
		TokenProvider:     "success-executable",
		TokenProviderType: OneShotProvider,
		TokenProviderArgs: []string{"arg1", "arg2"},
	}
	mockExecRunner := mockExecRunner{}
	mockCmd := mockCmd{}
	mockExecRunner.On("LookPath", config.TokenProvider).
		Return(config.TokenProvider, nil)
	mockExecRunner.On("CommandContext", mock.Anything,
		config.TokenProvider, config.TokenProviderArgs).
		Return(&mockCmd)
	mockCmd.On("Start").Return(nil)
	mockCmd.On("Wait").Return(nil)
	cancel, err := testCommon(config, &mockExecRunner)
	defer cancel()
	assert.Nil(t, err)
	mockExecRunner.AssertExpectations(t)
	mockCmd.AssertExpectations(t)
}

func TestFailure(t *testing.T) {
	config := secretstoreclient.SecretServiceInfo{
		TokenProvider:     "failure-executable",
		TokenProviderType: OneShotProvider,
		TokenProviderArgs: []string{"arg1", "arg2"},
	}
	mockExecRunner := mockExecRunner{}
	mockCmd := mockCmd{}
	mockExecRunner.On("LookPath", config.TokenProvider).
		Return(config.TokenProvider, nil)
	mockExecRunner.On("CommandContext", mock.Anything,
		config.TokenProvider, config.TokenProviderArgs).
		Return(&mockCmd)
	mockCmd.On("Start").Return(nil)
	mockCmd.On("Wait").Return(&exec.ExitError{ProcessState: &os.ProcessState{}})
	cancel, err := testCommon(config, &mockExecRunner)
	defer cancel()
	assert.Error(t, err)
	mockExecRunner.AssertExpectations(t)
	mockCmd.AssertExpectations(t)
}

func testCommon(config secretstoreclient.SecretServiceInfo, mockExecRunner ExecRunner) (context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())
	p := NewTokenProvider(ctx, logger.MockLogger{}, mockExecRunner)
	if err := p.SetConfiguration(config); err != nil {
		return cancel, err
	}
	return cancel, p.Launch()
}
