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

package secretstore

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

const OneShotProvider = "oneshot"

type TokenProvider struct {
	loggingClient logger.LoggingClient
	ctx           context.Context
	execRunner    ExecRunner
	initialized   bool
	secretStore   config.SecretStoreInfo
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
func (p *TokenProvider) SetConfiguration(secretStore config.SecretStoreInfo) error {
	var err error
	p.secretStore = secretStore
	if p.secretStore.TokenProviderType != OneShotProvider {
		err = fmt.Errorf("%s is not a supported TokenProviderType", p.secretStore.TokenProviderType)
		return err
	}
	resolvedPath, err := p.execRunner.LookPath(p.secretStore.TokenProvider)
	if err != nil {
		err = fmt.Errorf("Failed to locate %s on PATH: %s", p.secretStore.TokenProvider, err.Error())
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

	p.loggingClient.Infof(
		"Launching token provider %s with arguments %s",
		p.resolvedPath,
		strings.Join(p.secretStore.TokenProviderArgs, " "))

	cmd := p.execRunner.CommandContext(p.ctx, p.resolvedPath, p.secretStore.TokenProviderArgs...)
	if err := cmd.Start(); err != nil {
		// For example, this might occur if a shared library was missing
		err = fmt.Errorf("%s failed to launch: %s", p.resolvedPath, err.Error())
		return err
	}

	err := cmd.Wait()
	if exitError, ok := err.(*exec.ExitError); ok {
		waitStatus := exitError.Sys().(syscall.WaitStatus)
		err = fmt.Errorf("%s terminated with non-zero exit code %d", p.resolvedPath, waitStatus.ExitStatus())
		return err
	}
	if err != nil {
		err = fmt.Errorf("%s failed with unexpected error: %s", p.resolvedPath, err.Error())
		return err
	}

	p.loggingClient.Info("token provider exited successfully")
	return nil
}
