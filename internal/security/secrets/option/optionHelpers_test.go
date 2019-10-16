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

	"github.com/edgexfoundry/edgex-go/internal/security/secrets"
	"github.com/stretchr/testify/assert"
)

func TestGetWorkDir(t *testing.T) {
	// sets up configuration for SecretsSetupInfo
	tearDown := setupGenerateTest(t)
	defer tearDown()

	// test WorkDir that's configured by toml
	runTimeDir, err := getWorkDir()
	if err != nil {
		t.Errorf("Error getting workdir, %v", err)
	}
	expectedWorkDir, err := filepath.Abs("./workingtest")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, expectedWorkDir, runTimeDir)

	// test default WorkDir
	secrets.Configuration = &secrets.ConfigurationStruct{
		SecretsSetup: secrets.SecretsSetupInfo{
			WorkDir: "",
		},
	}
	runTimeDir, err = getWorkDir()
	assert.Nil(t, err)
	assert.Equal(t, filepath.Join(defaultWorkDir, pkiInitBaseDir), runTimeDir)

	// test env variable
	os.Setenv(envXdgRuntimeDir, "/run")
	runTimeDir, err = getWorkDir()
	assert.Nil(t, err)
	assert.Equal(t, filepath.Join("/run", pkiInitBaseDir), runTimeDir)
}

func TestGetCertConfigDir(t *testing.T) {
	// sets up configuration data for SecretsSetupInfo
	tearDown := setupGenerateTest(t)
	defer tearDown()

	// test CertConfigDir that's configured by toml
	certConfigDir, err := getCertConfigDir()
	assert.Nil(t, err)
	assert.Equal(t, "./res", certConfigDir)

	// certificate config dir not configured in toml
	secrets.Configuration = &secrets.ConfigurationStruct{
		SecretsSetup: secrets.SecretsSetupInfo{
			CertConfigDir: "",
		},
	}
	certConfigDir, err = getCertConfigDir()
	assert.NotNil(t, err)
	assert.Equal(t, "", certConfigDir)

	// certificate config dir is configured but does not exist
	secrets.Configuration = &secrets.ConfigurationStruct{
		SecretsSetup: secrets.SecretsSetupInfo{
			CertConfigDir: "./fakePath",
		},
	}
	_, err = getCertConfigDir()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestGetCacheDir(t *testing.T) {
	// sets up configuration data for SecretsSetupInfo
	tearDown := setupCacheTest(t)
	defer tearDown()

	// test CacheDir that's configured by toml
	cacheDir, err := getCacheDir()
	assert.Nil(t, err)
	assert.Equal(t, "./cachetest", cacheDir)

	// test default cacheDir
	secrets.Configuration = &secrets.ConfigurationStruct{
		SecretsSetup: secrets.SecretsSetupInfo{
			CacheDir: "",
		},
	}
	cacheDir, _ = getCacheDir()
	assert.Equal(t, defaultPkiCacheDir, cacheDir)

	// cache directory is configured but does not exist
	secrets.Configuration = &secrets.ConfigurationStruct{
		SecretsSetup: secrets.SecretsSetupInfo{
			CacheDir: "./fakePath",
		},
	}
	_, err = getCacheDir()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestGetDeployDir(t *testing.T) {
	tearDown := setupImportTest(t)
	defer tearDown()

	// test DeployDir that's configured by toml
	deployDir, err := getDeployDir()
	assert.Nil(t, err)
	assert.Equal(t, "./deploytest", deployDir)

	// test default DeployDir
	secrets.Configuration = &secrets.ConfigurationStruct{
		SecretsSetup: secrets.SecretsSetupInfo{
			DeployDir: "",
		},
	}
	deployDir, err = getDeployDir()
	assert.Nil(t, err)
	assert.Equal(t, defaultPkiDeployDir, deployDir)

	// deploy directory is configured but does not exist
	secrets.Configuration = &secrets.ConfigurationStruct{
		SecretsSetup: secrets.SecretsSetupInfo{
			DeployDir: "./fakepath",
		},
	}
	_, err = getDeployDir()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}
