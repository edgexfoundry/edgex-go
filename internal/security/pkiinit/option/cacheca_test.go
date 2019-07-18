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

	"github.com/stretchr/testify/assert"
)

func TestCacheCAOn(t *testing.T) {
	pkisetupLocal = true
	vaultJSONPkiSetupExist = true
	tearDown := setupCacheCATest(t)
	defer tearDown(t)

	options := PkiInitOption{
		CacheCAOpt: true,
	}
	cacheCAOn, _, _ := NewPkiInitOption(options)
	cacheCAOn.(*PkiInitOption).executor = testExecutor

	var exitStatus exitCode
	var err error
	f := CacheCA()
	exitStatus, err = f(cacheCAOn.(*PkiInitOption))

	cacheEmpty, emptyErr := isDirEmpty(getPkiCacheDirEnv())

	assert := assert.New(t)
	assert.Equal(normal, exitStatus)
	assert.Nil(err)
	assert.Nil(emptyErr)
	assert.False(cacheEmpty)

	generatedDirPath := filepath.Join(os.Getenv(envXdgRuntimeDir), pkiInitGeneratedDir)
	caPrivateKeyFile := filepath.Join(generatedDirPath, caServiceName, tlsSecretFileName)
	if _, checkErr := os.Stat(caPrivateKeyFile); os.IsNotExist(checkErr) {
		// didnt' find the file caPrivateKeyFile
		assert.Fail("CA private key does not exist!")
	}

	deployEmpty, emptyErr := isDirEmpty(pkiInitDeployDir)
	assert.Nil(emptyErr)
	assert.False(deployEmpty)
}

func TestCacheCACannotChangeError(t *testing.T) {
	// this test case is for calling cacheca after
	// a cache was called sucessfully previously
	pkisetupLocal = true
	vaultJSONPkiSetupExist = true
	tearDown := setupCacheCATest(t)
	defer tearDown(t)

	options := PkiInitOption{
		CacheOpt: true,
	}
	cacheOn, _, _ := NewPkiInitOption(options)
	cacheOn.(*PkiInitOption).executor = testExecutor

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

	// now we do the cacheca and expect to get error
	options = PkiInitOption{
		CacheCAOpt: true,
	}
	cacheCAOn, _, _ := NewPkiInitOption(options)
	cacheCAOn.(*PkiInitOption).executor = testExecutor

	f = CacheCA()
	exitStatus, err = f(cacheCAOn.(*PkiInitOption))

	cacheEmpty, emptyErr = isDirEmpty(getPkiCacheDirEnv())

	assert.Equal(exitWithError, exitStatus)
	assert.NotNil(err)
	assert.Nil(emptyErr)
	assert.False(cacheEmpty)
}

func TestCacheCAOff(t *testing.T) {
	pkisetupLocal = true
	vaultJSONPkiSetupExist = true
	tearDown := setupCacheCATest(t)
	defer tearDown(t)

	options := PkiInitOption{
		CacheCAOpt: false,
	}
	cacheCAOff, _, _ := NewPkiInitOption(options)
	cacheCAOff.(*PkiInitOption).executor = testExecutor
	exitCode, err := cacheCAOff.executeOptions(CacheCA())

	assert := assert.New(t)
	assert.Equal(normal, exitCode)
	assert.Nil(err)
}

func setupCacheCATest(t *testing.T) func(t *testing.T) {
	return setupCacheTest(t)
}
