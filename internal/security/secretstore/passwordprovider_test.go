//
// Copyright (c) 2021 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package secretstore

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
)

func TestInvalidPasswordProvider(t *testing.T) {
	serviceInfo := config.SecretStoreInfo{
		PasswordProvider: "does-not-exist",
	}
	mockExecRunner := &mockExecRunner{}
	mockExecRunner.On("LookPath", serviceInfo.PasswordProvider).
		Return("", errors.New("fake file does not exist"))
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := NewPasswordProvider(logger.MockLogger{}, mockExecRunner)
	err := p.SetConfiguration(serviceInfo.PasswordProvider, serviceInfo.PasswordProviderArgs)
	assert.Error(t, err)
	mockExecRunner.AssertExpectations(t)
}

func TestPasswordConfigNotInitialized(t *testing.T) {
	mockExecRunner := &mockExecRunner{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := NewPasswordProvider(logger.MockLogger{}, mockExecRunner)
	// SetConfiguration deliberately missing
	_, err := p.Generate(ctx)
	assert.Error(t, err)
	mockExecRunner.AssertExpectations(t)
}

func TestPasswordProviderFailsToStart(t *testing.T) {
	serviceInfo := config.SecretStoreInfo{
		PasswordProvider:     "failing-executable",
		PasswordProviderArgs: []string{"arg1", "arg2"},
	}
	mockExecRunner := &mockExecRunner{}
	mockCmd := mockCmd{}
	mockExecRunner.On("LookPath", serviceInfo.PasswordProvider).
		Return(serviceInfo.PasswordProvider, nil)
	mockExecRunner.On("SetStdout", mock.Anything).Once()
	mockExecRunner.On("CommandContext", mock.Anything,
		serviceInfo.PasswordProvider, serviceInfo.PasswordProviderArgs).
		Return(&mockCmd)
	mockCmd.On("Start").Return(errors.New("error starting"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := NewPasswordProvider(logger.MockLogger{}, mockExecRunner)
	err := p.SetConfiguration(serviceInfo.PasswordProvider, serviceInfo.PasswordProviderArgs)
	assert.NoError(t, err)
	_, err = p.Generate(ctx)
	assert.Error(t, err)
	mockExecRunner.AssertExpectations(t)
}

func TestPasswordProviderFailsAtRuntime(t *testing.T) {
	serviceInfo := config.SecretStoreInfo{
		PasswordProvider:     "failing-executable",
		PasswordProviderArgs: []string{"arg1", "arg2"},
	}
	mockExecRunner := &mockExecRunner{}
	mockCmd := mockCmd{}
	mockExecRunner.On("LookPath", serviceInfo.PasswordProvider).
		Return(serviceInfo.PasswordProvider, nil)
	mockExecRunner.On("SetStdout", mock.Anything).Once()
	mockExecRunner.On("CommandContext", mock.Anything,
		serviceInfo.PasswordProvider, serviceInfo.PasswordProviderArgs).
		Return(&mockCmd)
	mockCmd.On("Start").Return(nil)
	mockCmd.On("Wait").Return(&exec.ExitError{ProcessState: &os.ProcessState{}})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := NewPasswordProvider(logger.MockLogger{}, mockExecRunner)
	err := p.SetConfiguration(serviceInfo.PasswordProvider, serviceInfo.PasswordProviderArgs)
	assert.NoError(t, err)
	_, err = p.Generate(ctx)
	assert.Error(t, err)
	mockExecRunner.AssertExpectations(t)
}
