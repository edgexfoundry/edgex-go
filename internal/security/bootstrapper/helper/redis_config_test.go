/*******************************************************************************
 * Copyright 2023 Intel Corporation
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
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateRedisConfig(t *testing.T) {
	testConfFile := "testConfFile"
	testACLFile := "testACLFile"
	confFile, err := os.OpenFile(testConfFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	require.NoError(t, err)
	defer func() {
		_ = confFile.Close()
		_ = os.RemoveAll(testConfFile)
	}()

	maxClient := 1000

	err = GenerateRedisConfig(confFile, testACLFile, maxClient)
	require.NoError(t, err)

	inputFile, err := os.Open(testConfFile)
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

	require.Equal(t, 2, len(outputlines))
	require.Equal(t, "aclfile "+testACLFile, outputlines[0])
	require.Equal(t, fmt.Sprintf("maxclients %d", maxClient), outputlines[1])
}

func TestGenerateACLConfig(t *testing.T) {
	testACLFile := "testACLFile"
	aclFile, err := os.OpenFile(testACLFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	require.NoError(t, err)
	defer func() {
		_ = aclFile.Close()
		_ = os.RemoveAll(testACLFile)
	}()

	testFakePwd := "123456abcdefg!@#$%^&"

	err = GenerateACLConfig(aclFile, &testFakePwd)
	require.NoError(t, err)

	inputFile, err := os.Open(testACLFile)
	require.NoError(t, err)
	defer inputFile.Close()
	inputScanner := bufio.NewScanner(inputFile)
	inputScanner.Split(bufio.ScanLines)
	var outputlines []string
	// Read until a newline for each Scan
	for inputScanner.Scan() {
		line := inputScanner.Text()
		outputlines = append(outputlines, line)
	}

	require.Equal(t, 1, len(outputlines))
	require.Equal(t, fmt.Sprintf("user default on allkeys allchannels +@all -@dangerous #%x",
		sha256.Sum256([]byte(testFakePwd))), outputlines[0])
	// should not be equal if use different password
	require.NotEqual(t, fmt.Sprintf("user default on allkeys allchannels +@all -@dangerous #%x",
		sha256.Sum256([]byte("differentPassword"))), outputlines[0])
}
