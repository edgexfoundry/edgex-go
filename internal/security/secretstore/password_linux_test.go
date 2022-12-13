//go:build linux
// +build linux

//
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package secretstore

import (
	"context"
	"net/http"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateWithAPG(t *testing.T) {
	rootToken := "s.Ga5jyNq6kNfRMVQk2LY1j9iu" // nolint:gosec
	mockLogger := logger.MockLogger{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Note: apg only available with gnome-desktop, expected to be missing on server Linux distros
	gk := NewPasswordGenerator(mockLogger, "apg", []string{"-a", "1", "-n", "1", "-m", "12", "-x", "64"})
	cr := NewCred(&http.Client{}, rootToken, gk, "", logger.MockLogger{})

	p1, err := cr.GeneratePassword(ctx)
	require.NoError(t, err, "failed to create credential")
	p2, err := cr.GeneratePassword(ctx)
	require.NoError(t, err, "failed to create credential")
	assert.NotEqual(t, p1, p2, "each call to GeneratePassword should return a new password")
}
