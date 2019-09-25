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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/setup"
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
	f := Cache()
	exitStatus, err = f(cacheOn.(*PkiInitOption))

	cacheEmpty, emptyErr := isDirEmpty(getPkiCacheDirEnv())

	assert := assert.New(t)
	assert.Equal(normal, exitStatus)
	assert.Nil(err)
	assert.Nil(emptyErr)
	assert.False(cacheEmpty)

	generatedDirPath := filepath.Join(os.Getenv(envXdgRuntimeDir), pkiInitGeneratedDir)
	caPrivateKeyFile := filepath.Join(generatedDirPath, caServiceName, tlsSecretFileName)
	if _, checkErr := os.Stat(caPrivateKeyFile); checkErr == nil {
		// means found the file caPrivateKeyFile
		assert.Fail("CA private key are not removed!")
	}

	deployEmpty, emptyErr := isDirEmpty(pkiInitDeployDir)
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
	f := Cache()
	exitStatus, err = f(cacheOn.(*PkiInitOption))

	cacheEmpty, emptyErr := isDirEmpty(getPkiCacheDirEnv())

	assert := assert.New(t)
	assert.Equal(normal, exitStatus)
	assert.Nil(err)
	assert.Nil(emptyErr)
	assert.False(cacheEmpty)

	generatedDirPath := filepath.Join(os.Getenv(envXdgRuntimeDir), pkiInitGeneratedDir)
	// now we move the whole generated directory and leave the cache dir untouched
	os.RemoveAll(generatedDirPath)

	// call the cache option again:
	exitStatus, err = f(cacheOn.(*PkiInitOption))

	cacheEmpty, emptyErr = isDirEmpty(getPkiCacheDirEnv())

	assert.Equal(normal, exitStatus)
	assert.Nil(err)
	assert.Nil(emptyErr)
	assert.False(cacheEmpty)

	deployEmpty, emptyErr := isDirEmpty(pkiInitDeployDir)
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

	_ = setup.Init()

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

	origEnvXdgRuntimeDir := os.Getenv(envXdgRuntimeDir)
	// change it to the current working directory
	os.Setenv(envXdgRuntimeDir, curDir)

	origEnvPkiCache := os.Getenv(envPkiCache)
	// use curDir/cache as the working directory for test
	pkiCacheDir := filepath.Join(curDir, "cache")
	os.Setenv(envPkiCache, pkiCacheDir)
	_ = createDirectoryIfNotExists(pkiCacheDir)

	origDeployDir := pkiInitDeployDir
	tempDir, tempDirErr := ioutil.TempDir(curDir, "deploytest")
	if tempDirErr != nil {
		t.Fatalf("cannot create temporary scratch directory for the test: %v", tempDirErr)
	}
	pkiInitDeployDir = tempDir

	return func() {
		// cleanup
		os.Remove(jsonVaultFile)
		os.Remove(jsonKongFile)
		os.Setenv(envXdgRuntimeDir, origEnvXdgRuntimeDir)
		os.Setenv(envPkiCache, origEnvPkiCache)
		os.RemoveAll(pkiInitDeployDir)
		os.RemoveAll(pkiCacheDir)
		os.RemoveAll(testScratchDir)
		os.RemoveAll(testGeneratedDir)
		os.RemoveAll(testResourceDir)
		pkiInitScratchDir = origScratchDir
		pkiInitGeneratedDir = origGeneratedDir
		pkiInitDeployDir = origDeployDir
		vaultJSONPkiSetupExist = true
		kongJSONPkiSetupExist = true
	}
}
