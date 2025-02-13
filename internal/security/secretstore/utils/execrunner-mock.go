//
// Copyright (c) 2021 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package utils

import (
	"context"
	"io"

	"github.com/stretchr/testify/mock"
)

type MockExecRunner struct {
	mock.Mock
}

func (m *MockExecRunner) SetStdout(stdout io.Writer) {
	m.Called(stdout)
}

func (m *MockExecRunner) LookPath(file string) (string, error) {
	arguments := m.Called(file)
	return arguments.String(0), arguments.Error(1)
}

func (m *MockExecRunner) CommandContext(ctx context.Context,
	name string, arg ...string) CmdRunner {
	arguments := m.Called(ctx, name, arg)
	return arguments.Get(0).(CmdRunner)
}

type MockCmd struct {
	mock.Mock
}

func (m *MockCmd) Start() error {
	arguments := m.Called()
	return arguments.Error(0)
}

func (m *MockCmd) Wait() error {
	arguments := m.Called()
	return arguments.Error(0)
}
