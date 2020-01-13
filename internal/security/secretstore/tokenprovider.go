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
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

const OneShotProvider = "oneshot"

type ExecRunner interface {
	LookPath(file string) (string, error)
	CommandContext(ctx context.Context, name string, arg ...string) CmdRunner
}

type CmdRunner interface {
	Start() error
	Wait() error
}

type ExecWrapper struct{}

func (w ExecWrapper) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (w ExecWrapper) CommandContext(ctx context.Context, name string, arg ...string) CmdRunner {
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

type TokenProvider struct {
	loggingClient logger.LoggingClient
	ctx           context.Context
	execRunner    ExecRunner
	initialized   bool
	config        secretstoreclient.SecretServiceInfo
	resolvedPath  string
}

// NewTokenProvider creates a new TokenProvider
func NewTokenProvider(ctx context.Context, lc logger.LoggingClient, execRunner ExecRunner) *TokenProvider {
	return &TokenProvider{
		loggingClient: lc,
		ctx:           ctx,
		execRunner:    execRunner,
	}
}

// SetConfiguration parses token provider configuration and resolves paths specified therein
func (p *TokenProvider) SetConfiguration(config secretstoreclient.SecretServiceInfo) error {
	var err error
	p.config = config
	if p.config.TokenProviderType != OneShotProvider {
		err := fmt.Errorf("%s is not a supported TokenProviderType", p.config.TokenProviderType)
		p.loggingClient.Error(err.Error())
		return err
	}
	resolvedPath, err := p.execRunner.LookPath(p.config.TokenProvider)
	if err != nil {
		p.loggingClient.Error(fmt.Sprintf("Failed to locate %s on PATH: %s", p.config.TokenProvider, err.Error()))
		return err
	}
	p.initialized = true
	p.resolvedPath = resolvedPath
	return nil
}

// Launch spawns the token provider function
func (p *TokenProvider) Launch() error {
	if !p.initialized {
		err := fmt.Errorf("TokenProvider object not initialized; call SetConfiguration() first")
		return err
	}

	p.loggingClient.Info(fmt.Sprintf("Launching token provider %s with arguments %s", p.resolvedPath, strings.Join(p.config.TokenProviderArgs, " ")))
	cmd := p.execRunner.CommandContext(p.ctx, p.resolvedPath, p.config.TokenProviderArgs...)
	if err := cmd.Start(); err != nil {
		// For example, this might occur if a shared library was missing
		p.loggingClient.Error(fmt.Sprintf("%s failed to launch: %s", p.resolvedPath, err.Error()))
		return err
	}

	err := cmd.Wait()
	if exitError, ok := err.(*exec.ExitError); ok {
		waitStatus := exitError.Sys().(syscall.WaitStatus)
		p.loggingClient.Error(fmt.Sprintf("%s terminated with non-zero exit code %d", p.resolvedPath, waitStatus.ExitStatus()))
		return err
	}
	if err != nil {
		p.loggingClient.Error(fmt.Sprintf("%s failed with unexpected error: %s", p.resolvedPath, err.Error()))
		return err
	}

	p.loggingClient.Info("token provider exited successfully")
	return nil
}
