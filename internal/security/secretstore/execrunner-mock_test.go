//
// Copyright (c) 2021 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package secretstore

import (
	"context"
	"io"

	"github.com/stretchr/testify/mock"
)

type mockExecRunner struct {
	mock.Mock
}

func (m *mockExecRunner) SetStdout(stdout io.Writer) {
	m.Called(stdout)
}

func (m *mockExecRunner) LookPath(file string) (string, error) {
	arguments := m.Called(file)
	return arguments.String(0), arguments.Error(1)
}

func (m *mockExecRunner) CommandContext(ctx context.Context,
	name string, arg ...string) CmdRunner {
	arguments := m.Called(ctx, name, arg)
	return arguments.Get(0).(CmdRunner)
}

type mockCmd struct {
	mock.Mock
}

func (m *mockCmd) Start() error {
	arguments := m.Called()
	return arguments.Error(0)
}

func (m *mockCmd) Wait() error {
	arguments := m.Called()
	return arguments.Error(0)
}
