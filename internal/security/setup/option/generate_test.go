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
	defer tearDown()

	options := PkiInitOption{
		GenerateOpt: true,
	}

	assert := assert.New(t)

	generateOn, _, _ := NewPkiInitOption(options)
	generateOn.(*PkiInitOption).executor = testExecutor

	f := Generate()
	exitCode, err := f(generateOn.(*PkiInitOption))
	assert.Equal(normal, exitCode)
	assert.Nil(err)

	workDir, err := getWorkDir()
	assert.Nil(err)

	generatedDirPath := filepath.Join(workDir, pkiInitGeneratedDir)
	caPrivateKeyFile := filepath.Join(generatedDirPath, caServiceName, tlsSecretFileName)
	fileExists := checkIfFileExists(caPrivateKeyFile)
	assert.False(fileExists, "CA private key are not removed!")

	deployDir, err := getDeployDir()
	assert.Nil(err)

	deployEmpty, emptyErr := isDirEmpty(deployDir)
	assert.Nil(emptyErr)
	assert.False(deployEmpty)

	// check sentinel file is present when deploy is done
	sentinel := filepath.Join(deployDir, caServiceName, pkiInitFilePerServiceComplete)
	fileExists = checkIfFileExists(sentinel)
	assert.True(fileExists, "sentinel file does not exist for CA service!")

	sentinel = filepath.Join(deployDir, vaultServiceName, pkiInitFilePerServiceComplete)
	fileExists = checkIfFileExists(sentinel)
	assert.True(fileExists, "sentinel file does not exist for vault service!")
}

func TestGenerateWithVaultJSONPkiSetupMissing(t *testing.T) {
	vaultJSONPkiSetupExist = false // this will lead to missing json
	kongJSONPkiSetupExist = true
	tearDown := setupGenerateTest(t)
	defer tearDown()

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
	defer tearDown()

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
	defer tearDown()

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

func setupGenerateTest(t *testing.T) func() {
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

	err = setup.Init()
	if err != nil {
		t.Fatalf("Failed to init security-secrets-setup: %v", err)
	}

	origEnvVal := os.Getenv(envXdgRuntimeDir)
	os.Unsetenv(envXdgRuntimeDir) // unset env var, so it uses the config toml
	oldConfig := setup.Configuration

	testWorkDir, err := getWorkDir()
	if err != nil {
		t.Fatalf("Error getting work dir for the test: %v", err)
	}

	testScratchDir := filepath.Join(testWorkDir, pkiInitScratchDir)
	if err := createDirectoryIfNotExists(testScratchDir); err != nil {
		t.Fatalf("cannot create scratch dir %s for the test: %v", testScratchDir, err)
	}

	testGeneratedDir := filepath.Join(testWorkDir, pkiInitGeneratedDir)
	if err := createDirectoryIfNotExists(testGeneratedDir); err != nil {
		t.Fatalf("cannot create generated dir %s for the test: %v", testGeneratedDir, err)
	}

	pkiInitDeployDir := "./deploytest"
	if err := createDirectoryIfNotExists(pkiInitDeployDir); err != nil {
		t.Fatalf("cannot create deploy dir %s for the test: %v", pkiInitDeployDir, err)
	}

	return func() {
		// cleanup
		os.Remove(jsonVaultFile)
		os.Remove(jsonKongFile)
		os.RemoveAll(testWorkDir)
		os.RemoveAll(testScratchDir)
		os.RemoveAll(testGeneratedDir)
		os.RemoveAll(pkiInitDeployDir)
		os.RemoveAll(testResourceDir)
		os.Setenv(envXdgRuntimeDir, origEnvVal)
		setup.Configuration = oldConfig
		vaultJSONPkiSetupExist = true
		kongJSONPkiSetupExist = true
	}
}
