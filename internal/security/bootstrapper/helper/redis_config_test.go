/*******************************************************************************
 * Copyright 2021 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package helper

import (
	"bufio"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateConfig(t *testing.T) {
	testConfFile := "testConfFile"
	confFile, err := os.OpenFile(testConfFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	require.NoError(t, err)
	defer func() {
		err = confFile.Close()
		require.NoError(t, err)
		err = os.RemoveAll(testConfFile)
		require.NoError(t, err)
	}()

	fw := bufio.NewWriter(confFile)
	testFakePwd := "123456abcdefg!@#$%^&"

	err = GenerateConfig(fw, &testFakePwd)
	require.NoError(t, err)
	require.NoError(t, fw.Flush())

	inputFile, err := os.Open(testConfFile)
	require.NoError(t, err)
	defer inputFile.Close()
	inputScanner := bufio.NewScanner(inputFile)
	inputScanner.Split(bufio.ScanLines)
	lineCount := 0
	// Read until a newline for each Scan
	for inputScanner.Scan() {
		lineCount++
		line := inputScanner.Text()
		require.Contains(t, line, testFakePwd)
	}
	require.Equal(t, 2, lineCount)
}
