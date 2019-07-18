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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testExecutor *mockOptionsExecutor
var pkisetupLocal bool
var vaultJSONPkiSetupExist bool

func TestGenerate(t *testing.T) {
	pkisetupLocal = true
	vaultJSONPkiSetupExist = true
	tearDown := setupGenerateTest(t)
	defer tearDown(t)

	options := PkiInitOption{
		GenerateOpt: true,
	}
	generateOn, _, _ := NewPkiInitOption(options)
	generateOn.(*PkiInitOption).executor = testExecutor

	f := Generate()
	exitCode, err := f(generateOn.(*PkiInitOption))

	assert := assert.New(t)
	assert.Equal(normal, exitCode)
	assert.Nil(err)
	generatedDirPath := filepath.Join(os.Getenv(envXdgRuntimeDir), pkiInitGeneratedDir)
	caPrivateKeyFile := filepath.Join(generatedDirPath, caServiceName, tlsSecretFileName)
	fileExists := checkIfFileExists(caPrivateKeyFile)
	assert.False(fileExists, "CA private key are not removed!")

	deployEmpty, emptyErr := isDirEmpty(pkiInitDeployDir)
	assert.Nil(emptyErr)
	assert.False(deployEmpty)

	// check sentinel file is present when deploy is done
	sentinel := filepath.Join(pkiInitDeployDir, caServiceName, pkiInitFilePerServiceComplete)
	fileExists = checkIfFileExists(sentinel)
	assert.True(fileExists, "sentinel file does not exist for CA service!")

	sentinel = filepath.Join(pkiInitDeployDir, vaultServiceName, pkiInitFilePerServiceComplete)
	fileExists = checkIfFileExists(sentinel)
	assert.True(fileExists, "sentinel file does not exist for vault service!")
}

func TestGenerateWithPkiSetupMissing(t *testing.T) {
	pkisetupLocal = false // this will lead to pkisetup binary missing
	vaultJSONPkiSetupExist = true
	tearDown := setupGenerateTest(t)
	defer tearDown(t)

	options := PkiInitOption{
		GenerateOpt: true,
	}
	generateOn, _, _ := NewPkiInitOption(options)
	generateOn.(*PkiInitOption).executor = testExecutor

	f := Generate()
	exitCode, err := f(generateOn.(*PkiInitOption))

	assert := assert.New(t)
	assert.Equal(exitWithError, exitCode)
	assert.NotNil(err)
}

func TestGenerateWithVaultJSONPkiSetupMissing(t *testing.T) {
	pkisetupLocal = true
	vaultJSONPkiSetupExist = false // this will lead to missing json
	tearDown := setupGenerateTest(t)
	defer tearDown(t)

	options := PkiInitOption{
		GenerateOpt: true,
	}
	generateOn, _, _ := NewPkiInitOption(options)
	generateOn.(*PkiInitOption).executor = testExecutor

	f := Generate()
	exitCode, err := f(generateOn.(*PkiInitOption))

	assert := assert.New(t)
	assert.Equal(exitWithError, exitCode)
	assert.NotNil(err)
}

func TestGenerateOff(t *testing.T) {
	pkisetupLocal = true
	vaultJSONPkiSetupExist = true
	tearDown := setupGenerateTest(t)
	defer tearDown(t)

	options := PkiInitOption{
		GenerateOpt: false,
	}
	generateOff, _, _ := NewPkiInitOption(options)
	generateOff.(*PkiInitOption).executor = testExecutor
	exitCode, err := generateOff.executeOptions(Generate())

	assert := assert.New(t)
	assert.Equal(normal, exitCode)
	assert.Nil(err)
}

func setupGenerateTest(t *testing.T) func(t *testing.T) {
	testExecutor = &mockOptionsExecutor{}
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get the working dir %s: %v", curDir, err)
	}

	pkiSetupFile := filepath.Join(curDir, pkiSetupExecutable)
	if pkisetupLocal {
		if _, err := copyFile(filepath.Join(curDir, "..", "..", "..", "..", "cmd", "pkisetup", pkiSetupExecutable), pkiSetupFile); err != nil {
			t.Fatalf("cannot copy pkisetup binary for the test: %v", err)
		}
		os.Chmod(pkiSetupFile, 0777)
	}

	jsonVaultFile := filepath.Join(curDir, pkiSetupVaultJSON)
	if vaultJSONPkiSetupExist {
		if _, err := copyFile(filepath.Join(curDir, "..", "..", "..", "..", "configs", pkiSetupVaultJSON), jsonVaultFile); err != nil {
			t.Fatalf("cannot copy %s for the test: %v", pkiSetupVaultJSON, err)
		}
	}

	origEnvXdgRuntimeDir := os.Getenv(envXdgRuntimeDir)
	fmt.Println("Env XDG_RUNTIME_DIR: ", origEnvXdgRuntimeDir)

	// change it to the current working directory
	os.Setenv(envXdgRuntimeDir, curDir)

	origScratchDir := pkiInitScratchDir
	testScratchDir, tempDirErr := ioutil.TempDir(curDir, "scratch")
	if tempDirErr != nil {
		t.Fatalf("cannot create temporary scratch directory for the test: %v", tempDirErr)
	}
	pkiInitScratchDir = filepath.Base(testScratchDir)

	origGeneratedDir := pkiInitGeneratedDir
	testGeneratedDir, tempDirErr := ioutil.TempDir(curDir, "generated")
	if tempDirErr != nil {
		t.Fatalf("cannot create temporary generated directory for the test: %v", tempDirErr)
	}
	pkiInitGeneratedDir = filepath.Base(testGeneratedDir)

	origDeployDir := pkiInitDeployDir
	tempDir, tempDirErr := ioutil.TempDir(curDir, "deploytest")
	if tempDirErr != nil {
		t.Fatalf("cannot create temporary scratch directory for the test: %v", tempDirErr)
	}
	pkiInitDeployDir = tempDir

	return func(t *testing.T) {
		// cleanup
		os.Remove(pkiSetupFile)
		os.Remove(jsonVaultFile)
		os.Setenv(envXdgRuntimeDir, origEnvXdgRuntimeDir)
		os.RemoveAll(testScratchDir)
		os.RemoveAll(testGeneratedDir)
		os.RemoveAll(pkiInitDeployDir)
		pkiInitScratchDir = origScratchDir
		pkiInitGeneratedDir = origGeneratedDir
		pkiInitDeployDir = origDeployDir
		pkisetupLocal = true
		vaultJSONPkiSetupExist = true
	}
}
