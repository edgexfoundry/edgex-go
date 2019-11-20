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

package _import

import (
	"io/ioutil"
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

func TestImportCacheDirEmpty(t *testing.T) {
	// this test case is for import running first
	// and PKI_CACHE_DIR is empty- should expect an error by design
	tearDown, configuration := setupImportTest(t)
	defer tearDown()

	command, _ := NewCommand(logger.NewMockClient(), configuration)
	exitCode, err := command.Execute()

	assert.NotNil(t, err)
	assert.Equal(t, contract.StatusCodeExitWithError, exitCode)

	deployDir, err := helper.GetDeployDir(configuration)
	assert.Nil(t, err)

	deployEmpty, emptyErr := helper.IsDirEmpty(deployDir)
	assert.Nil(t, emptyErr)
	assert.True(t, deployEmpty)
}

func TestImportFromPKICache(t *testing.T) {
	// this test case is for import pre-populated cached PKI
	// files from PKI_CACHE dir
	tearDown, configuration := setupImportTest(t)
	defer tearDown()

	// put some test file into the cache dir first
	writeTestFileToCacheDir(t, configuration)

	command, _ := NewCommand(logger.NewMockClient(), configuration)
	exitCode, err := command.Execute()

	assert.Equal(t, contract.StatusCodeExitNormal, exitCode)
	assert.Nil(t, err)

	deployDir, err := helper.GetDeployDir(configuration)
	assert.Nil(t, err)

	deployEmpty, emptyErr := helper.IsDirEmpty(deployDir)
	assert.Nil(t, emptyErr)
	assert.False(t, deployEmpty)
}

func TestEmptyPkiCacheEnvironment(t *testing.T) {
	command, _ := NewCommand(logger.NewMockClient(), getConfiguration())
	exitCode, err := command.Execute()

	// when PKI_CACHE env is empty, it leads to non-existing dir
	// and should be an error case
	assert.NotNil(t, err)
	assert.Equal(t, contract.StatusCodeExitWithError, exitCode)
}

func TestImportOff(t *testing.T) {
	tearDown, configuration := setupImportTest(t)
	defer tearDown()

	// put some test file into the cache dir first
	writeTestFileToCacheDir(t, configuration)

	command, _ := NewCommand(logger.NewMockClient(), configuration)
	exitCode, err := command.Execute()

	assert.Equal(t, contract.StatusCodeExitNormal, exitCode)
	assert.Nil(t, err)
}

func getConfiguration() *config.ConfigurationStruct {
	return &config.ConfigurationStruct{
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
}

func setupImportTest(t *testing.T) (func(), *config.ConfigurationStruct) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get the working dir %s: %v", curDir, err)
	}

	testResourceDir := filepath.Join(curDir, test.ResourceDirName)
	if err := helper.CreateDirectoryIfNotExists(testResourceDir); err != nil {
		t.Fatalf("cannot create resource dir %s for the test: %v", testResourceDir, err)
	}

	resDir := filepath.Join(curDir, "testdata", test.ResourceDirName)

	testResTomlFile := filepath.Join(testResourceDir, test.ConfigTomlFile)
	if _, err := helper.CopyFile(filepath.Join(resDir, test.ConfigTomlFile), testResTomlFile); err != nil {
		t.Fatalf("cannot copy %s for the test: %v", test.ConfigTomlFile, err)
	}

	origEnvXdgRuntimeDir := os.Getenv(helper.EnvXdgRuntimeDir)
	// change it to the current working directory
	os.Setenv(helper.EnvXdgRuntimeDir, curDir)

	pkiCacheDir := "./cachetest"
	if err := helper.CreateDirectoryIfNotExists(pkiCacheDir); err != nil {
		t.Fatalf("cannot create the PKI_CACHE dir %s: %v", pkiCacheDir, err)
	}

	pkiInitDeployDir := "./deploytest"
	if err := helper.CreateDirectoryIfNotExists(pkiInitDeployDir); err != nil {
		t.Fatalf("cannot create dir %s for the test: %v", pkiInitDeployDir, err)
	}

	return func() {
		// cleanup
		os.Setenv(helper.EnvXdgRuntimeDir, origEnvXdgRuntimeDir)
		os.RemoveAll(pkiInitDeployDir)
		os.RemoveAll(pkiCacheDir)
		os.RemoveAll(testResourceDir)
	}, getConfiguration()
}

func writeTestFileToCacheDir(t *testing.T, configuration *config.ConfigurationStruct) {
	pkiCacheDir, err := helper.GetCacheDir(configuration)
	if err != nil {
		t.Fatalf("Cache directory %s error: %v", pkiCacheDir, err)
	}
	// make a test dir
	testFileDir := filepath.Join(pkiCacheDir, "test", generate.CaServiceName)
	_ = helper.CreateDirectoryIfNotExists(testFileDir)
	testFile := filepath.Join(testFileDir, "testFile")
	testData := []byte("test data\n")
	if err := ioutil.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("cannot write testData to directory %s: %v", pkiCacheDir, err)
	}
}
