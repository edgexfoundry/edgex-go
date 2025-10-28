/*******************************************************************************
 * Copyright 2021 Intel Corporation
 * Copyright (C) 2024 IOTech Ltd
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
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"

	"github.com/stretchr/testify/require"
)

func TestMarkComplete(t *testing.T) {
	testDir := "testDir"
	doneFile := "testDone"

	defer (cleanupDir(testDir))()

	err := os.MkdirAll(testDir, os.ModePerm) //#nosec G301 -- Directory intentionally world-accessible for shared access
	require.NoError(t, err)

	err = MarkComplete(testDir, doneFile)
	require.NoError(t, err)

	require.FileExists(t, filepath.Join(testDir, doneFile))
}

func TestCreateDirectoryIfNotExists(t *testing.T) {
	testDir := "testDirNew"
	require.NoDirExists(t, testDir)
	defer (cleanupDir(testDir))()

	err := CreateDirectoryIfNotExists(testDir)
	require.NoError(t, err)
	require.DirExists(t, testDir)

	testPreCreatedDir := "pre-created-dir"
	defer (cleanupDir(testPreCreatedDir))()

	err = os.MkdirAll(testPreCreatedDir, os.ModePerm) //#nosec G301 -- Directory intentionally world-accessible for shared access
	require.NoError(t, err)
	err = CreateDirectoryIfNotExists(testPreCreatedDir)
	require.NoError(t, err)
	require.DirExists(t, testPreCreatedDir)
}

// cleanupDir deletes all files in the directory and files in the directory
func cleanupDir(dir string) func() {
	return func() {
		_ = os.RemoveAll(dir)
	}
}

func TestCreateConfigFile(t *testing.T) {
	lc := logger.NewMockClient()

	testDir := "testDirNew"
	require.NoDirExists(t, testDir)
	defer (cleanupDir(testDir))()

	err := CreateDirectoryIfNotExists(testDir)
	require.NoError(t, err)
	require.DirExists(t, testDir)

	// Define the file path for the test
	testFilePath := "testfile.txt"

	// Call the function to create the file
	_, err = CreateConfigFile(testDir, testFilePath, lc)
	require.NoError(t, err)

	require.FileExists(t, path.Join(testDir, testFilePath))
}
