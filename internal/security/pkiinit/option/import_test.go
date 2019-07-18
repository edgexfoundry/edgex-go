//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0'
//
package option

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestImportPriorFileChange(t *testing.T) {
	// this test case is for import running first
	// and waiting for files from PKI_CACHE dir to be appeared

	tearDown := setupImportTest(t)
	defer tearDown(t)

	options := PkiInitOption{
		ImportOpt: true,
	}
	importOn, _, _ := NewPkiInitOption(options)
	importOn.(*PkiInitOption).executor = testExecutor

	var exitStatus exitCode
	var err error
	go func() { // in a gorountine, so it won't block
		f := Import()
		exitStatus, err = f(importOn.(*PkiInitOption))
	}()

	// to allow time for go func() to be properly initialized
	time.Sleep(time.Second)

	// now put a testFile into cache dir
	writeTestFileToCacheDir(t)

	// to allow time to finish deploy in another go-routine
	time.Sleep(4 * time.Second)

	deployEmpty, emptyErr := isDirEmpty(pkiInitDeployDir)

	assert := assert.New(t)
	assert.Equal(normal, exitStatus)
	assert.Nil(err)
	assert.Nil(emptyErr)
	assert.False(deployEmpty)
}

func TestImportPostFileChange(t *testing.T) {
	// this test case is for import running later
	// files from PKI_CACHE dir are already in-place before import is called

	tearDown := setupImportTest(t)
	defer tearDown(t)

	// put some test file into the cache dir first
	writeTestFileToCacheDir(t)

	options := PkiInitOption{
		ImportOpt: true,
	}
	importOn, _, _ := NewPkiInitOption(options)
	importOn.(*PkiInitOption).executor = testExecutor

	f := Import()

	exitStatus, err := f(importOn.(*PkiInitOption))

	deployEmpty, emptyErr := isDirEmpty(pkiInitDeployDir)

	assert := assert.New(t)
	assert.Equal(normal, exitStatus)
	assert.Nil(err)
	assert.Nil(emptyErr)
	assert.False(deployEmpty)
}

func TestEmptyPkiCacheEnvironment(t *testing.T) {
	options := PkiInitOption{
		ImportOpt: true,
	}
	importOn, _, _ := NewPkiInitOption(options)
	importOn.(*PkiInitOption).executor = testExecutor
	exitCode, err := importOn.executeOptions(Import())

	// when PKI_CACHE env is empty, it leads to non-existing dir
	// and should be an error case
	assert := assert.New(t)
	assert.NotNil(err)
	assert.Equal(exitWithError, exitCode)
}

func TestImportOff(t *testing.T) {
	tearDown := setupImportTest(t)
	defer tearDown(t)

	options := PkiInitOption{
		ImportOpt: false,
	}
	importOff, _, _ := NewPkiInitOption(options)
	importOff.(*PkiInitOption).executor = testExecutor
	exitCode, err := importOff.executeOptions(Import())

	assert := assert.New(t)
	assert.Equal(normal, exitCode)
	assert.Nil(err)
}

func TestIsDirEmpty(t *testing.T) {
	assert := assert.New(t)
	_, err := isDirEmpty("/non/existing/dir/")

	assert.NotNil(err)

	// put some test file into the current dir to trigger event
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get the working dir %s: %v", curDir, err)
	}
	empty, err := isDirEmpty(curDir)
	assert.Nil(err)
	assert.False(empty)

	// create an empty temp dir
	tempDir, err := ioutil.TempDir(curDir, "test")
	if err != nil {
		t.Fatalf("cannot create the temporary dir %s: %v", tempDir, err)
	}
	empty, err = isDirEmpty(tempDir)
	defer func() {
		// remove tempDir:
		os.RemoveAll(tempDir)
	}()

	assert.Nil(err)
	assert.True(empty)
}

func setupImportTest(t *testing.T) func(t *testing.T) {
	testExecutor = &mockOptionsExecutor{}
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get the working dir %s: %v", curDir, err)
	}

	origEnvXdgRuntimeDir := os.Getenv(envXdgRuntimeDir)
	// change it to the current working directory
	os.Setenv(envXdgRuntimeDir, curDir)

	origEnvPkiCache := os.Getenv(envPkiCache)
	// use curDir/cache as the working directory for test
	pkiCacheDir := filepath.Join(curDir, "cache")
	os.Setenv(envPkiCache, pkiCacheDir)
	createDirectoryIfNotExists(pkiCacheDir)

	origDeployDir := pkiInitDeployDir
	tempDir, tempDirErr := ioutil.TempDir(curDir, "deploytest")
	if tempDirErr != nil {
		t.Fatalf("cannot create temporary scratch directory for the test: %v", tempDirErr)
	}
	pkiInitDeployDir = tempDir

	return func(t *testing.T) {
		// cleanup
		os.Setenv(envXdgRuntimeDir, origEnvXdgRuntimeDir)
		os.Setenv(envPkiCache, origEnvPkiCache)
		os.RemoveAll(pkiInitDeployDir)
		os.RemoveAll(pkiCacheDir)
		pkiInitDeployDir = origDeployDir
	}
}

func writeTestFileToCacheDir(t *testing.T) {
	pkiCacheDir := os.Getenv(envPkiCache)
	// make a test dir
	testFileDir := filepath.Join(pkiCacheDir, "test", caServiceName)
	_ = createDirectoryIfNotExists(testFileDir)
	testFile := filepath.Join(testFileDir, "testFile")
	testData := []byte("test data\n")
	if err := ioutil.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("cannot write testData to direcotry %s: %v", pkiCacheDir, err)
	}
}
