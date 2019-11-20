//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0'
//

package generate

import (
	"os"
	"path/filepath"
	"testing"

	types "github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/contract"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/helper"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/test"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	tearDown, configuration := SetupGenerateTest(t, test.VaultJSONPkiSetupExists, test.KongJSONPkiSetupExists)
	defer tearDown()

	command, _ := NewCommand(logger.NewMockClient(), configuration)
	exitCode, err := command.Execute()

	assert.Equal(t, contract.StatusCodeExitNormal, exitCode)
	assert.Nil(t, err)

	workDir, err := helper.GetWorkDir(configuration)
	assert.Nil(t, err)

	generatedDirPath := filepath.Join(workDir, PkiInitGeneratedDir)
	caPrivateKeyFile := filepath.Join(generatedDirPath, CaServiceName, TlsSecretFileName)
	fileExists := helper.CheckIfFileExists(caPrivateKeyFile)
	assert.False(t, fileExists, "CA private key are not removed!")

	deployDir, err := helper.GetDeployDir(configuration)
	assert.Nil(t, err)

	deployEmpty, emptyErr := helper.IsDirEmpty(deployDir)
	assert.Nil(t, emptyErr)
	assert.False(t, deployEmpty)

	// check sentinel file is present when Deploy is done
	sentinel := filepath.Join(deployDir, CaServiceName, helper.PkiInitFilePerServiceComplete)
	fileExists = helper.CheckIfFileExists(sentinel)
	assert.True(t, fileExists, "sentinel file does not exist for CA service!")

	sentinel = filepath.Join(deployDir, vaultServiceName, helper.PkiInitFilePerServiceComplete)
	fileExists = helper.CheckIfFileExists(sentinel)
	assert.True(t, fileExists, "sentinel file does not exist for vault service!")
}

func TestGenerateWithVaultJSONPkiSetupMissing(t *testing.T) {
	tearDown, configuration := SetupGenerateTest(t, test.VaultJSONPkiSetupDoesNotExist, test.KongJSONPkiSetupExists)
	defer tearDown()

	command, _ := NewCommand(logger.NewMockClient(), configuration)
	exitCode, err := command.Execute()

	assert.Equal(t, contract.StatusCodeExitWithError, exitCode)
	assert.NotNil(t, err)
}

func TestGenerateWithKongJSONPkiSetupMissing(t *testing.T) {
	tearDown, configuration := SetupGenerateTest(t, test.VaultJSONPkiSetupExists, test.KongJSONPkiSetupDoesNotExist)
	defer tearDown()

	command, _ := NewCommand(logger.NewMockClient(), configuration)
	exitCode, err := command.Execute()

	assert.Equal(t, contract.StatusCodeExitWithError, exitCode)
	assert.NotNil(t, err)
}

func SetupGenerateTest(t *testing.T, vaultJSONPkiSetupExist bool, kongJSONPkiSetupExist bool) (func(), *config.ConfigurationStruct) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get the working dir %s: %v", curDir, err)
	}

	testResourceDir := filepath.Join(curDir, test.ResourceDirName)
	if err := helper.CreateDirectoryIfNotExists(testResourceDir); err != nil {
		t.Fatalf("cannot create resource dir %s for the test: %v", testResourceDir, err)
	}

	resDir := filepath.Join(curDir, "testdata", test.ResourceDirName)
	jsonVaultFile := filepath.Join(curDir, test.ResourceDirName, PkiSetupVaultJSON)
	if vaultJSONPkiSetupExist {
		if _, err := helper.CopyFile(filepath.Join(resDir, PkiSetupVaultJSON), jsonVaultFile); err != nil {
			t.Fatalf("cannot copy %s for the test: %v", PkiSetupVaultJSON, err)
		}
	}

	jsonKongFile := filepath.Join(curDir, test.ResourceDirName, PkiSetupKongJSON)
	if kongJSONPkiSetupExist {
		if _, err := helper.CopyFile(filepath.Join(resDir, PkiSetupKongJSON), jsonKongFile); err != nil {
			t.Fatalf("cannot copy %s for the test: %v", PkiSetupKongJSON, err)
		}
	}

	testResTomlFile := filepath.Join(testResourceDir, test.ConfigTomlFile)
	if _, err := helper.CopyFile(filepath.Join(resDir, test.ConfigTomlFile), testResTomlFile); err != nil {
		t.Fatalf("cannot copy %s for the test: %v", test.ConfigTomlFile, err)
	}

	configuration := &config.ConfigurationStruct{
		Writable: config.WritableInfo{
			LogLevel: "DEBUG",
		},
		Logging: types.LoggingInfo{
			EnableRemote: false,
			File:         "./logs/security-secrets-setup.log",
		},
		SecretsSetup: config.SecretsSetupInfo{
			WorkDir:       "./workingtest",
			DeployDir:     "./deploytest",
			CacheDir:      "./cachetest",
			CertConfigDir: "./res",
		},
	}

	origEnvVal := os.Getenv(helper.EnvXdgRuntimeDir)
	os.Unsetenv(helper.EnvXdgRuntimeDir) // unset env var, so it uses the config toml

	testWorkDir, err := helper.GetWorkDir(configuration)
	if err != nil {
		t.Fatalf("Error getting work dir for the test: %v", err)
	}

	testScratchDir := filepath.Join(testWorkDir, PkiInitScratchDir)
	if err := helper.CreateDirectoryIfNotExists(testScratchDir); err != nil {
		t.Fatalf("cannot create scratch dir %s for the test: %v", testScratchDir, err)
	}

	testGeneratedDir := filepath.Join(testWorkDir, PkiInitGeneratedDir)
	if err := helper.CreateDirectoryIfNotExists(testGeneratedDir); err != nil {
		t.Fatalf("cannot create generated dir %s for the test: %v", testGeneratedDir, err)
	}

	pkiInitDeployDir := "./deploytest"
	if err := helper.CreateDirectoryIfNotExists(pkiInitDeployDir); err != nil {
		t.Fatalf("cannot create Deploy dir %s for the test: %v", pkiInitDeployDir, err)
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
		os.Setenv(helper.EnvXdgRuntimeDir, origEnvVal)
	}, configuration
}
