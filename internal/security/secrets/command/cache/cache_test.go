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
package cache

import (
	"os"
	"path/filepath"
	"testing"

	types "github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/command/generate"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/contract"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/helper"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/test"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/stretchr/testify/assert"
)

func TestCacheOn(t *testing.T) {
	tearDown, configuration := setupCacheTest(t, test.VaultJSONPkiSetupExists, test.KongJSONPkiSetupExists)
	defer tearDown()

	loggerMock := logger.NewMockClient()
	generateCommand, _ := generate.NewCommand(generate.NewFlags(), loggerMock, configuration)
	command, _ := NewCommand(NewFlags(), loggerMock, configuration, generateCommand)
	exitCode, err := command.Execute()

	assert.Nil(t, err)
	assert.Equal(t, contract.StatusCodeExitNormal, exitCode)

	cacheDir, err := helper.GetCacheDir(configuration)
	assert.Nil(t, err)

	cacheEmpty, emptyErr := helper.IsDirEmpty(cacheDir)
	assert.Nil(t, emptyErr)
	assert.False(t, cacheEmpty)

	workDir, err := helper.GetWorkDir(configuration)
	if err != nil {
		t.Errorf("Error getting workdir, %v", err)
	}
	generatedDirPath := filepath.Join(workDir, generate.PkiInitGeneratedDir)
	caPrivateKeyFile := filepath.Join(generatedDirPath, generate.CaServiceName, generate.TlsSecretFileName)
	if _, checkErr := os.Stat(caPrivateKeyFile); checkErr == nil {
		// means found the file caPrivateKeyFile
		assert.Fail(t, "CA private key are not removed!")
	}

	deployDir, err := helper.GetDeployDir(configuration)
	assert.Nil(t, err)

	deployEmpty, emptyErr := helper.IsDirEmpty(deployDir)
	assert.Nil(t, emptyErr)
	assert.False(t, deployEmpty)
}

func TestCacheDirNotEmpty(t *testing.T) {
	tearDown, configuration := setupCacheTest(t, test.VaultJSONPkiSetupExists, test.KongJSONPkiSetupExists)
	defer tearDown()

	loggerMock := logger.NewMockClient()
	generateCommand, _ := generate.NewCommand(generate.NewFlags(), loggerMock, configuration)
	command, _ := NewCommand(NewFlags(), loggerMock, configuration, generateCommand)
	exitCode, err := command.Execute()

	assert.Nil(t, err)
	assert.Equal(t, contract.StatusCodeExitNormal, exitCode)

	cacheDir, err := helper.GetCacheDir(configuration)
	assert.Nil(t, err)

	cacheEmpty, emptyErr := helper.IsDirEmpty(cacheDir)
	assert.Nil(t, emptyErr)
	assert.False(t, cacheEmpty)

	workDir, err := helper.GetWorkDir(configuration)
	assert.Nil(t, err)

	generatedDirPath := filepath.Join(workDir, generate.PkiInitGeneratedDir)
	// now we move the whole generated directory and leave the cache dir untouched
	os.RemoveAll(generatedDirPath)

	// call the cache option again:
	exitCode, err = command.Execute()
	assert.Nil(t, err)
	assert.Equal(t, contract.StatusCodeExitNormal, exitCode)

	cacheEmpty, emptyErr = helper.IsDirEmpty(cacheDir)
	assert.Nil(t, emptyErr)
	assert.False(t, cacheEmpty)

	deployDir, err := helper.GetDeployDir(configuration)
	assert.Nil(t, err)

	deployEmpty, emptyErr := helper.IsDirEmpty(deployDir)
	assert.Nil(t, emptyErr)
	assert.False(t, deployEmpty)
}

func TestCacheOff(t *testing.T) {
	tearDown, configuration := setupCacheTest(t, test.VaultJSONPkiSetupExists, test.KongJSONPkiSetupExists)
	defer tearDown()

	loggerMock := logger.NewMockClient()
	generateCommand, _ := generate.NewCommand(generate.NewFlags(), loggerMock, configuration)
	command, _ := NewCommand(NewFlags(), loggerMock, configuration, generateCommand)
	exitCode, err := command.Execute()

	assert.Equal(t, contract.StatusCodeExitNormal, exitCode)
	assert.Nil(t, err)
}

func setupCacheTest(t *testing.T, vaultJSONPkiSetupExist bool, kongJSONPkiSetupExist bool) (func(), *config.ConfigurationStruct) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get the working dir %s: %v", curDir, err)
	}

	testResourceDir := filepath.Join(curDir, test.ResourceDirName)
	if err := helper.CreateDirectoryIfNotExists(testResourceDir); err != nil {
		t.Fatalf("cannot create resource dir %s for the test: %v", testResourceDir, err)
	}

	resDir := filepath.Join(curDir, "testdata", test.ResourceDirName)
	jsonVaultFile := filepath.Join(curDir, test.ResourceDirName, generate.PkiSetupVaultJSON)
	if vaultJSONPkiSetupExist {
		if _, err := helper.CopyFile(filepath.Join(resDir, generate.PkiSetupVaultJSON), jsonVaultFile); err != nil {
			t.Fatalf("cannot copy %s for the test: %v", generate.PkiSetupVaultJSON, err)
		}
	}

	jsonKongFile := filepath.Join(curDir, test.ResourceDirName, generate.PkiSetupKongJSON)
	if kongJSONPkiSetupExist {
		if _, err := helper.CopyFile(filepath.Join(resDir, generate.PkiSetupKongJSON), jsonKongFile); err != nil {
			t.Fatalf("cannot copy %s for the test: %v", generate.PkiSetupKongJSON, err)
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

	os.Unsetenv(helper.EnvXdgRuntimeDir) // unset env var, so it uses the config toml

	testWorkDir, err := helper.GetWorkDir(configuration)
	if err != nil {
		t.Fatalf("Error getting work dir for the test: %v", err)
	}

	testScratchDir := filepath.Join(testWorkDir, generate.PkiInitScratchDir)
	if err := helper.CreateDirectoryIfNotExists(testScratchDir); err != nil {
		t.Fatalf("cannot create scratch dir %s for the test: %v", testScratchDir, err)
	}

	testGeneratedDir := filepath.Join(testWorkDir, generate.PkiInitGeneratedDir)
	if err := helper.CreateDirectoryIfNotExists(testGeneratedDir); err != nil {
		t.Fatalf("cannot create generated dir %s for the test: %v", testGeneratedDir, err)
	}

	pkiCacheDir := "./cachetest"
	if err := helper.CreateDirectoryIfNotExists(pkiCacheDir); err != nil {
		t.Fatalf("cannot create cache dir %s for the test: %v", pkiCacheDir, err)
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
		os.RemoveAll(pkiCacheDir)
		os.RemoveAll(testScratchDir)
		os.RemoveAll(testGeneratedDir)
		os.RemoveAll(pkiInitDeployDir)
		os.RemoveAll(testResourceDir)
	}, configuration
}
