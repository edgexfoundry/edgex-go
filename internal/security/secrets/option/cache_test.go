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

func TestCacheOn(t *testing.T) {
	vaultJSONPkiSetupExist = true
	kongJSONPkiSetupExist = true
	tearDown := setupCacheTest(t)
	defer tearDown()

	options := PkiInitOption{
		CacheOpt: true,
	}
	cacheOn, _, _ := NewPkiInitOption(options)
	cacheOn.setExecutor(testExecutor)

	var exitStatus exitCode
	var err error
	assert := assert.New(t)

	f := Cache()
	exitStatus, err = f(cacheOn.(*PkiInitOption))
	assert.Nil(err)
	assert.Equal(normal, exitStatus)

	cacheDir, err := getCacheDir()
	assert.Nil(err)

	cacheEmpty, emptyErr := isDirEmpty(cacheDir)
	assert.Nil(emptyErr)
	assert.False(cacheEmpty)

	workDir, err := getWorkDir()
	if err != nil {
		t.Errorf("Error getting workdir, %v", err)
	}
	generatedDirPath := filepath.Join(workDir, pkiInitGeneratedDir)
	caPrivateKeyFile := filepath.Join(generatedDirPath, caServiceName, tlsSecretFileName)
	if _, checkErr := os.Stat(caPrivateKeyFile); checkErr == nil {
		// means found the file caPrivateKeyFile
		assert.Fail("CA private key are not removed!")
	}

	deployDir, err := getDeployDir()
	assert.Nil(err)

	deployEmpty, emptyErr := isDirEmpty(deployDir)
	assert.Nil(emptyErr)
	assert.False(deployEmpty)
}

func TestCacheDirNotEmpty(t *testing.T) {
	vaultJSONPkiSetupExist = true
	kongJSONPkiSetupExist = true
	tearDown := setupCacheTest(t)
	defer tearDown()

	options := PkiInitOption{
		CacheOpt: true,
	}
	cacheOn, _, _ := NewPkiInitOption(options)
	cacheOn.setExecutor(testExecutor)

	var exitStatus exitCode
	var err error
	assert := assert.New(t)

	f := Cache()
	exitStatus, err = f(cacheOn.(*PkiInitOption))
	assert.Nil(err)
	assert.Equal(normal, exitStatus)

	cacheDir, err := getCacheDir()
	assert.Nil(err)

	cacheEmpty, emptyErr := isDirEmpty(cacheDir)
	assert.Nil(emptyErr)
	assert.False(cacheEmpty)

	workDir, err := getWorkDir()
	assert.Nil(err)

	generatedDirPath := filepath.Join(workDir, pkiInitGeneratedDir)
	// now we move the whole generated directory and leave the cache dir untouched
	os.RemoveAll(generatedDirPath)

	// call the cache option again:
	exitStatus, err = f(cacheOn.(*PkiInitOption))
	assert.Nil(err)
	assert.Equal(normal, exitStatus)

	cacheEmpty, emptyErr = isDirEmpty(cacheDir)
	assert.Nil(emptyErr)
	assert.False(cacheEmpty)

	deployDir, err := getDeployDir()
	assert.Nil(err)

	deployEmpty, emptyErr := isDirEmpty(deployDir)
	assert.Nil(emptyErr)
	assert.False(deployEmpty)
}

func TestCacheOff(t *testing.T) {
	vaultJSONPkiSetupExist = true
	kongJSONPkiSetupExist = true
	tearDown := setupCacheTest(t)
	defer tearDown()

	options := PkiInitOption{
		CacheOpt: false,
	}
	cacheOff, _, _ := NewPkiInitOption(options)
	cacheOff.setExecutor(testExecutor)
	exitCode, err := cacheOff.executeOptions(Cache())

	assert := assert.New(t)
	assert.Equal(normal, exitCode)
	assert.Nil(err)
}

func setupCacheTest(t *testing.T) func() {
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

	err = secrets.Init("")
	if err != nil {
		t.Fatalf("Failed to init security-secrets-setup: %v", err)
	}

	os.Unsetenv(envXdgRuntimeDir) // unset env var, so it uses the config toml
	oldConfig := secrets.Configuration

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

	pkiCacheDir := "./cachetest"
	if err := createDirectoryIfNotExists(pkiCacheDir); err != nil {
		t.Fatalf("cannot create cache dir %s for the test: %v", pkiCacheDir, err)
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
		os.RemoveAll(pkiCacheDir)
		os.RemoveAll(testScratchDir)
		os.RemoveAll(testGeneratedDir)
		os.RemoveAll(pkiInitDeployDir)
		os.RemoveAll(testResourceDir)
		secrets.Configuration = oldConfig
		vaultJSONPkiSetupExist = true
		kongJSONPkiSetupExist = true
	}
}
