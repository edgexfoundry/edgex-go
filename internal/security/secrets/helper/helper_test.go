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

package helper

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/secrets/config"

	"github.com/stretchr/testify/assert"
)

func envSetup() func() {
	envOrig := os.Getenv(EnvXdgRuntimeDir)
	os.Unsetenv(EnvXdgRuntimeDir)
	return func() {
		os.Setenv(EnvXdgRuntimeDir, envOrig)
	}
}

func TestGetWorkDirValid(t *testing.T) {
	defer (envSetup())()
	workDir := "./workingtest"
	expectedWorkDir, _ := filepath.Abs(workDir)
	configuration := &config.ConfigurationStruct{
		SecretsSetup: config.SecretsSetupInfo{
			WorkDir: workDir,
		},
	}

	d, err := GetWorkDir(configuration)

	assert.Nil(t, err)
	assert.Equal(t, expectedWorkDir, d)
}

func TestGetWorkDirDefault(t *testing.T) {
	defer (envSetup())()
	configuration := &config.ConfigurationStruct{
		SecretsSetup: config.SecretsSetupInfo{
			WorkDir: "",
		},
	}

	d, err := GetWorkDir(configuration)

	assert.Nil(t, err)
	assert.Equal(t, filepath.Join(defaultWorkDir, pkiInitBaseDir), d)
}

func TestWorkDirEnvVar(t *testing.T) {
	defer (envSetup())()
	const workDir = "./tmp"
	os.Setenv(EnvXdgRuntimeDir, workDir)
	configuration := &config.ConfigurationStruct{}

	d, err := GetWorkDir(configuration)

	assert.Nil(t, err)
	assert.Equal(t, filepath.Join(workDir, pkiInitBaseDir), d)

}

func TestGetCertConfigDirValid(t *testing.T) {
	const testFileName = "test.file"
	certConfigDir, _ := filepath.Abs("./validDirectory")
	configuration := &config.ConfigurationStruct{
		SecretsSetup: config.SecretsSetupInfo{
			CertConfigDir: certConfigDir,
		},
	}
	if err := CreateDirectoryIfNotExists(certConfigDir); err != nil {
		assert.Fail(t, "unable to create directory")
	}
	defer func() {
		os.RemoveAll(certConfigDir)
	}()
	if _, err := os.Create(filepath.Join(certConfigDir, testFileName)); err != nil {
		assert.Fail(t, "unable to create file")
	}

	d, err := GetCertConfigDir(configuration)

	assert.Nil(t, err)
	assert.Equal(t, certConfigDir, d)
}

func TestGetCertConfigDirEmpty(t *testing.T) {
	configuration := &config.ConfigurationStruct{}

	_, err := GetCertConfigDir(configuration)

	assert.NotNil(t, err)
	assert.Equal(t, errMessageCertConfigDirNotSet, err.Error())
}

func TestGetCertConfigDirInvalid(t *testing.T) {
	configuration := &config.ConfigurationStruct{
		SecretsSetup: config.SecretsSetupInfo{
			CertConfigDir: "./directoryDoesNotExist",
		},
	}

	_, err := GetCertConfigDir(configuration)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestGetCacheDirValid(t *testing.T) {
	cacheDir, _ := filepath.Abs("./cachetest")
	configuration := &config.ConfigurationStruct{
		SecretsSetup: config.SecretsSetupInfo{
			CacheDir: cacheDir,
		},
	}
	if err := CreateDirectoryIfNotExists(cacheDir); err != nil {
		assert.Fail(t, "unable to create directory")
	}
	defer func() {
		os.Remove(cacheDir)
	}()

	d, err := GetCacheDir(configuration)

	assert.Nil(t, err)
	assert.Equal(t, cacheDir, d)
}

func TestGetCacheDirEmpty(t *testing.T) {
	configuration := &config.ConfigurationStruct{}

	d, err := GetCacheDir(configuration)

	assert.Nil(t, err)
	assert.Equal(t, defaultPkiCacheDir, d)
}

func TestGetCacheDirInvalid(t *testing.T) {
	configuration := &config.ConfigurationStruct{
		SecretsSetup: config.SecretsSetupInfo{
			CacheDir: "./directoryDoesNotExist",
		},
	}

	_, err := GetCacheDir(configuration)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestGetDeployDirValid(t *testing.T) {
	deployDir, _ := filepath.Abs("./deploytest")
	configuration := &config.ConfigurationStruct{
		SecretsSetup: config.SecretsSetupInfo{
			DeployDir: deployDir,
		},
	}
	if err := CreateDirectoryIfNotExists(deployDir); err != nil {
		assert.Fail(t, "unable to create directory")
	}
	defer func() {
		os.Remove(deployDir)
	}()

	d, err := GetDeployDir(configuration)

	assert.Nil(t, err)
	assert.Equal(t, deployDir, d)
}

func TestGetDeployDirEmpty(t *testing.T) {
	configuration := &config.ConfigurationStruct{}

	d, err := GetDeployDir(configuration)

	assert.Nil(t, err)
	assert.Equal(t, defaultPkiDeployDir, d)
}

func TestGetDeployDirInvalid(t *testing.T) {
	configuration := &config.ConfigurationStruct{
		SecretsSetup: config.SecretsSetupInfo{
			DeployDir: "./directoryDoesNotExist",
		},
	}

	_, err := GetDeployDir(configuration)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}
