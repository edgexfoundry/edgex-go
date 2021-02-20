//
// Copyright (c) 2021 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package secretstore

import (
	"context"
	"io"
	"os"
	"os/exec"
)

// CmdRunner is mockable interface for golang's exec.Cmd
type CmdRunner interface {
	Start() error
	Wait() error
}

// ExecRunner is mockable interface for wrapping os/exec functionality
type ExecRunner interface {
	SetStdout(stdout io.Writer)
	LookPath(file string) (string, error)
	CommandContext(ctx context.Context, name string, arg ...string) CmdRunner
}

type execWrapper struct {
	Stdout io.Writer
	Stderr io.Writer
}

// NewDefaultExecRunner creates an os/exec wrapper
// that joins subprocesses' stdout and stderr with the caller's
func NewDefaultExecRunner() ExecRunner {
	return &execWrapper{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// SetStdout allows overriding of stdout capture (for consuming password generator output)
func (w *execWrapper) SetStdout(stdout io.Writer) {
	w.Stdout = stdout
}

// LookPath wraps os/exec.LookPath
func (w *execWrapper) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// CommandContext wraps os/exec.CommandContext
func (w *execWrapper) CommandContext(ctx context.Context, name string, arg ...string) CmdRunner {
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Stdout = w.Stdout
	cmd.Stderr = w.Stderr
	return cmd
}
