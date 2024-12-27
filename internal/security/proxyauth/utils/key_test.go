//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"

	"github.com/stretchr/testify/require"
)

func TestSigningKeyName(t *testing.T) {
	mockIssuer := "mockIssuer"
	expected := mockIssuer + "/" + common.SigningKeyType
	result := SigningKeyName(mockIssuer)
	require.Equal(t, expected, result)
}

func TestVerificationKeyName(t *testing.T) {
	mockIssuer := "mockIssuer"
	expected := mockIssuer + "/" + common.VerificationKeyType
	result := VerificationKeyName(mockIssuer)
	require.Equal(t, expected, result)
}
