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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarkComplete(t *testing.T) {
	testDir := "testDir"
	doneFile := "testDone"

	defer (cleanupDir(testDir))()

	err := os.MkdirAll(testDir, os.ModePerm)
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

	err = os.MkdirAll(testPreCreatedDir, os.ModePerm)
	require.NoError(t, err)
	err = CreateDirectoryIfNotExists(testPreCreatedDir)
	require.NoError(t, err)
	require.DirExists(t, testPreCreatedDir)
}

func TestChownDirRecursively(t *testing.T) {
	testDir := "testChownDir"
	testFile := "testChownFile"

	defer (cleanupDir(testDir))()

	err := os.MkdirAll(testDir, os.ModePerm)
	require.NoError(t, err)

	testFilePath := filepath.Join(testDir, testFile)

	err = ioutil.WriteFile(testFilePath, []byte("this is a test"), 0600)
	require.NoError(t, err)

	err = ChownDirRecursive(testDir, os.Geteuid(), os.Getegid())
	require.NoError(t, err)

	// if try to change ownership to a different uid, since this test most likely will be run
	// as non-root user, so the error will be expected
	err = ChownDirRecursive(testDir, 999, os.Getegid())
	if os.Geteuid() != 0 {
		// expected permission issue since it is not allowed to change ownership to other user
		require.Error(t, err)
	} else {
		require.NoError(t, err)
	}
}

// cleanupDir deletes all files in the directory and files in the directory
func cleanupDir(dir string) func() {
	return func() {
		_ = os.RemoveAll(dir)
	}
}
