//
// Copyright (c) 2021 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package secretstore

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

type PasswordProvider struct {
	loggingClient        logger.LoggingClient
	execRunner           ExecRunner
	initialized          bool
	passwordProvider     string
	passwordProviderArgs []string
	resolvedPath         string
}

// NewPasswordProvider creates a new PasswordProvider
func NewPasswordProvider(lc logger.LoggingClient, execRunner ExecRunner) *PasswordProvider {
	return &PasswordProvider{
		loggingClient: lc,
		execRunner:    execRunner,
	}
}

// SetConfiguration parses token provider configuration and resolves paths specified therein
func (p *PasswordProvider) SetConfiguration(passwordProvider string, passwordProviderArgs []string) error {
	var err error
	p.passwordProvider = passwordProvider
	p.passwordProviderArgs = passwordProviderArgs
	resolvedPath, err := p.execRunner.LookPath(p.passwordProvider)
	if err != nil {
		p.loggingClient.Error(fmt.Sprintf("Failed to locate %s on PATH: %s", p.passwordProvider, err.Error()))
		return err
	}
	p.initialized = true
	p.resolvedPath = resolvedPath
	return nil
}

// Generate retrieves the password from the tool
func (p *PasswordProvider) Generate(ctx context.Context) (string, error) {
	var outputBuffer bytes.Buffer

	if !p.initialized {
		err := fmt.Errorf("PasswordProvider object not initialized; call SetConfiguration() first")
		return "", err
	}

	p.execRunner.SetStdout(&outputBuffer)

	p.loggingClient.Infof("Launching password provider %s with arguments %s", p.resolvedPath, strings.Join(p.passwordProviderArgs, " "))
	cmd := p.execRunner.CommandContext(ctx, p.resolvedPath, p.passwordProviderArgs...)
	if err := cmd.Start(); err != nil {
		// For example, this might occur if a shared library was missing
		p.loggingClient.Error(fmt.Sprintf("%s failed to launch: %s", p.resolvedPath, err.Error()))
		return "", err
	}

	err := cmd.Wait()
	if exitError, ok := err.(*exec.ExitError); ok {
		waitStatus := exitError.Sys().(syscall.WaitStatus)
		p.loggingClient.Error(fmt.Sprintf("%s terminated with non-zero exit code %d", p.resolvedPath, waitStatus.ExitStatus()))
		return "", err
	}
	if err != nil {
		p.loggingClient.Error(fmt.Sprintf("%s failed with unexpected error: %s", p.resolvedPath, err.Error()))
		return "", err
	}

	p.loggingClient.Info("password provider exited successfully")

	pw := outputBuffer.String()
	pw = strings.TrimSuffix(pw, "\n")
	return pw, nil
}
