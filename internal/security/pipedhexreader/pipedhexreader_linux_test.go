//go:build linux
// +build linux

//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package pipedhexreader

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPipedHexReaderNewline(t *testing.T) {
	// Arrange
	phr := NewPipedHexReader()
	expected, _ := hex.DecodeString("12345678")

	// Act
	key, err := phr.ReadHexBytesFromExe("./testdata/echowithnewline")

	// Assert
	require.NoError(t, err)
	require.Equal(t, expected, key)
}

func TestPipedHexReaderNoNewline(t *testing.T) {
	// Arrange
	phr := NewPipedHexReader()
	expected, _ := hex.DecodeString("12345678")

	// Act
	key, err := phr.ReadHexBytesFromExe("./testdata/echowithoutnewline")

	// Assert
	require.NoError(t, err)
	require.Equal(t, expected, key)
}
