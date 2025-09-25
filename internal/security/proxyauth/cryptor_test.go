//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package proxyauth

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaultAESKey(t *testing.T) {
	mockReadFile := func(path string) ([]byte, error) {
		if path == defaultAESKeyFile {
			return []byte("file-aes-key\n"), nil
		}
		return nil, errors.New("unexpected path")
	}

	tests := []struct {
		name         string
		mockEnvValue string
		mockReadFile func(string) ([]byte, error)
		expectKey    string
		expectErr    bool
	}{
		{
			name:         "load from environment variable",
			mockEnvValue: "env-aes-key",
			expectKey:    "env-aes-key",
		},
		{
			name:         "load from file",
			mockEnvValue: "",
			mockReadFile: mockReadFile,
			expectKey:    "file-aes-key",
		},
		{
			name:         "file not found",
			mockEnvValue: "",
			mockReadFile: func(path string) ([]byte, error) {
				return nil, os.ErrNotExist
			},
			expectErr: true,
		},
		{
			name:         "env key overrides file",
			mockEnvValue: "env-aes-key",
			mockReadFile: mockReadFile,
			expectKey:    "env-aes-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origGetenv := getenv
			origReadFile := readFile
			defer func() {
				getenv = origGetenv
				readFile = origReadFile
			}()

			getenv = func(k string) string {
				if k == EnvAesKey {
					return tt.mockEnvValue
				}
				return ""
			}

			if tt.mockReadFile != nil {
				readFile = tt.mockReadFile
			} else {
				readFile = origReadFile
			}

			key, err := loadDefaultAESKey()

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, key)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectKey, string(key))
		})
	}
}
