//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package helper

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneratePostgresScript(t *testing.T) {
	fileName := "testScriptFile"
	testScriptFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	require.NoError(t, err)
	defer func() {
		_ = testScriptFile.Close()
		_ = os.RemoveAll(fileName)
	}()

	mockUsername := "core_data"
	mockPassword := "password123"
	mockCredMap := []map[string]any{{
		"Username": mockUsername,
		"Password": &mockPassword,
	}}

	err = GeneratePostgresScript(testScriptFile, mockCredMap)
	require.NoError(t, err)

	inputFile, err := os.Open(fileName)
	require.NoError(t, err)
	defer inputFile.Close()

	inputScanner := bufio.NewScanner(inputFile)
	inputScanner.Split(bufio.ScanLines)
	var outputlines []string
	// Read until a newline for each Scan
	for inputScanner.Scan() {
		line := inputScanner.Text()
		if len(strings.TrimSpace(line)) > 0 { // only take non-empty line
			outputlines = append(outputlines, line)
		}
	}

	expectedCreateScript := fmt.Sprintf("CREATE USER %s with PASSWORD '%s';", mockUsername, mockPassword)
	require.Equal(t, 17, len(outputlines))
	require.Equal(t, expectedCreateScript, strings.TrimSpace(outputlines[11]))
}
