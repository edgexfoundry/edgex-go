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

	"github.com/edgexfoundry/edgex-go/internal/security/setup"
	"github.com/stretchr/testify/assert"
)

var testExecutor *mockOptionsExecutor
var vaultJSONPkiSetupExist bool
var kongJSONPkiSetupExist bool

func TestGenerate(t *testing.T) {
	vaultJSONPkiSetupExist = true
	kongJSONPkiSetupExist = true
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
	generatedDirPath := filepath.Join(getXdgRuntimeDir(), pkiInitGeneratedDir)
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

func TestGenerateWithVaultJSONPkiSetupMissing(t *testing.T) {
	vaultJSONPkiSetupExist = false // this will lead to missing json
	kongJSONPkiSetupExist = true
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

func TestGenerateWithKongJSONPkiSetupMissing(t *testing.T) {
	kongJSONPkiSetupExist = false // this will lead to missing json
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

func TestGenerateOff(t *testing.T) {
	vaultJSONPkiSetupExist = true
	kongJSONPkiSetupExist = true
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

func TestGetXdgRuntimeDir(t *testing.T) {
	origEnvVal := os.Getenv(envXdgRuntimeDir)
	os.Unsetenv(envXdgRuntimeDir)
	runTimeDir := getXdgRuntimeDir()
	assert.Equal(t, defaultXdgRuntimeDir, runTimeDir)

	os.Setenv(envXdgRuntimeDir, origEnvVal)
	runTimeDir = getXdgRuntimeDir()
	assert.Equal(t, origEnvVal, runTimeDir)
}

func setupGenerateTest(t *testing.T) func(t *testing.T) {
	testExecutor = &mockOptionsExecutor{}

	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get the working dir %s: %v", curDir, err)
	}

	testResourceDir := filepath.Join(curDir, resourceDirName)
	if err := createDirectoryIfNotExists(testResourceDir); err != nil {
		t.Fatalf("cannot create resource dir %s for the test: %v", testResourceDir, err)
	}

	resDir := filepath.Join(curDir, "testdata", resourceDirName)
	jsonVaultFile := filepath.Join(curDir, resourceDirName, pkiSetupVaultJSON)
	if vaultJSONPkiSetupExist {
		if _, err := copyFile(filepath.Join(resDir, pkiSetupVaultJSON), jsonVaultFile); err != nil {
			t.Fatalf("cannot copy %s for the test: %v", pkiSetupVaultJSON, err)
		}
	}

	jsonKongFile := filepath.Join(curDir, resourceDirName, pkiSetupKongJSON)
	if kongJSONPkiSetupExist {
		if _, err := copyFile(filepath.Join(resDir, pkiSetupKongJSON), jsonKongFile); err != nil {
			t.Fatalf("cannot copy %s for the test: %v", pkiSetupKongJSON, err)
		}
	}

	testResTomlFile := filepath.Join(testResourceDir, configTomlFile)
	if _, err := copyFile(filepath.Join(resDir, configTomlFile), testResTomlFile); err != nil {
		t.Fatalf("cannot copy %s for the test: %v", configTomlFile, err)
	}

	setup.Init()

	origEnvXdgRuntimeDir := getXdgRuntimeDir()
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
		os.Remove(jsonVaultFile)
		os.Remove(jsonKongFile)
		os.Setenv(envXdgRuntimeDir, origEnvXdgRuntimeDir)
		os.RemoveAll(testScratchDir)
		os.RemoveAll(testGeneratedDir)
		os.RemoveAll(pkiInitDeployDir)
		os.RemoveAll(testResourceDir)
		pkiInitScratchDir = origScratchDir
		pkiInitGeneratedDir = origGeneratedDir
		pkiInitDeployDir = origDeployDir
		vaultJSONPkiSetupExist = true
		kongJSONPkiSetupExist = true
	}
}
