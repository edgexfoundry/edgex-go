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

func TestGetWorkDirValid(t *testing.T) {
	workDir, _ := filepath.Abs("./workingtest")
	configuration := &config.ConfigurationStruct{
		SecretsSetup: config.SecretsSetupInfo{
			WorkDir: workDir,
		},
	}

	d, err := GetWorkDir(configuration)

	if err != nil {
		assert.Fail(t, "Error GetWorkDir, %v", err)
	}
	assert.Nil(t, err)
	assert.Equal(t, workDir, d)
}

func TestGetWorkDirDefault(t *testing.T) {
	configuration := &config.ConfigurationStruct{}

	d, err := GetWorkDir(configuration)

	assert.Nil(t, err)
	assert.Equal(t, filepath.Join(defaultWorkDir, pkiInitBaseDir), d)
}

func TestWorkDirEnvVar(t *testing.T) {
	const workDir = "./run"
	envOrig := os.Getenv(EnvXdgRuntimeDir)
	defer func() {
		os.Setenv(EnvXdgRuntimeDir, envOrig)
	}()
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
	if _, err := os.Create(filepath.Join(certConfigDir, testFileName)); err != nil {
		assert.Fail(t, "unable to create file")
	}
	defer func() {
		os.Remove(certConfigDir)
	}()

	d, err := GetCertConfigDir(configuration)

	if err != nil {
		assert.Fail(t, "Error GetCertConfigDir, %v", err)
	}
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

	if err != nil {
		assert.Fail(t, "Error GetCacheDir, %v", err)
	}
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

	if err != nil {
		assert.Fail(t, "Error GetDeployDir, %v", err)
	}
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
