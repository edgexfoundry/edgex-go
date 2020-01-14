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

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/secrets"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/command/generate"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/contract"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/helper"

	"github.com/stretchr/testify/assert"
)

func TestMainWithNoOption(t *testing.T) {
	defer (setupTest([]string{"cmd"}))()
	ctx, cancel := context.WithCancel(context.Background())

	configuration, exitStatusCode := secrets.Main(ctx, cancel)

	removeTestDirectories(configuration)
	assert.Equal(t, contract.StatusCodeExitNormal, exitStatusCode)
}

func TestMainWithConfigFileOption(t *testing.T) {
	defer (setupTest([]string{"cmd", "legacy", "-config", "./res/pkisetup-vault.json"}))()
	ctx, cancel := context.WithCancel(context.Background())

	configuration, exitStatusCode := secrets.Main(ctx, cancel)

	removeTestDirectories(configuration)
	assert.Equal(t, contract.StatusCodeExitNormal, exitStatusCode)

	// verify ./config/pki/EdgeXFoundryCA directory exists
	exists, err := doesDirectoryExist("./config/pki/EdgeXFoundryCA")
	if !exists {
		assert.NotNil(t, err)
		assert.FailNow(t, "cannot find the directory for TLS assets", err)
	}
}

func TestConfigFileOptionError(t *testing.T) {
	defer (setupTest([]string{"cmd", "legacy", "-config", "./non-exist/cert.json"}))()
	ctx, cancel := context.WithCancel(context.Background())

	configuration, exitStatusCode := secrets.Main(ctx, cancel)

	removeTestDirectories(configuration)
	assert.Equal(t, contract.StatusCodeExitWithError, exitStatusCode)
}

func TestMainWithGenerateOption(t *testing.T) {
	defer (setupTest([]string{"cmd", "--confdir=./testdata/res", "generate"}))()
	// following dir must match SecretsSetup.DeployDir value in configuration.toml
	d, _ := filepath.Abs("./deploytest")
	if err := helper.CreateDirectoryIfNotExists(d); err != nil {
		assert.Fail(t, "unable to create deploy directory")
	}
	ctx, cancel := context.WithCancel(context.Background())

	configuration, exitStatusCode := secrets.Main(ctx, cancel)

	removeTestDirectories(configuration)
	assert.Equal(t, contract.StatusCodeExitNormal, exitStatusCode)
}

func TestMainUnsupportedArgument(t *testing.T) {
	defer (setupTest([]string{"cmd", "unsupported"}))()
	ctx, cancel := context.WithCancel(context.Background())

	configuration, exitStatusCode := secrets.Main(ctx, cancel)

	removeTestDirectories(configuration)
	assert.Equal(t, contract.StatusCodeNoOptionSelected, exitStatusCode)
}

func TestMainVerifyMultipleSubcommands(t *testing.T) {
	defer (setupTest([]string{"cmd", "generate", "legacy"}))()
	ctx, cancel := context.WithCancel(context.Background())

	configuration, exitStatusCode := secrets.Main(ctx, cancel)

	removeTestDirectories(configuration)
	assert.Equal(t, contract.StatusCodeExitWithError, exitStatusCode)
}

func TestMainLegacySubcommandWithExtraArgs(t *testing.T) {
	defer (setupTest([]string{"cmd", "legacy", "-c", "./res/pkisetup-vault.json", "extra"}))()
	ctx, cancel := context.WithCancel(context.Background())

	configuration, exitStatusCode := secrets.Main(ctx, cancel)

	removeTestDirectories(configuration)
	assert.Equal(t, contract.StatusCodeExitWithError, exitStatusCode)
}

func TestMainWithCacheOption(t *testing.T) {
	defer (setupTest([]string{"cmd", "--confdir=./testdata/res", "cache"}))()
	// following dir must match SecretsSetup.DeployDir value in configuration.toml
	d, _ := filepath.Abs("./deploytest")
	if err := helper.CreateDirectoryIfNotExists(d); err != nil {
		assert.Fail(t, "unable to create deploy directory")
	}
	// must match SecretsSetup.CacheDir value in configuration.toml
	d, _ = filepath.Abs("./cachetest")
	if err := helper.CreateDirectoryIfNotExists(d); err != nil {
		assert.Fail(t, "unable to create cache directory")
	}
	ctx, cancel := context.WithCancel(context.Background())

	configuration, exitStatusCode := secrets.Main(ctx, cancel)

	removeTestDirectories(configuration)
	assert.Equal(t, contract.StatusCodeExitNormal, exitStatusCode)
}

func writeTestFileToCacheDir(t *testing.T, pkiCacheDir string) {
	testFileDir := filepath.Join(pkiCacheDir, "test", generate.CaServiceName)
	_ = helper.CreateDirectoryIfNotExists(testFileDir)
	testFile := filepath.Join(testFileDir, "testFile")
	testData := []byte("test data\n")
	if err := ioutil.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("cannot write testData to directory %s: %v", pkiCacheDir, err)
	}
}

func TestMainWithImportOption(t *testing.T) {
	defer (setupTest([]string{"cmd", "--confdir=./testdata/res", "import"}))()

	// following dir must match SecretsSetup.DeployDir value in configuration.toml
	d, _ := filepath.Abs("./deploytest")
	if err := helper.CreateDirectoryIfNotExists(d); err != nil {
		assert.Fail(t, "unable to create deploy directory")
	}
	d, _ = filepath.Abs("./cachetest")
	writeTestFileToCacheDir(t, d) // must match SecretsSetup.CacheDir value in configuration.toml
	ctx, cancel := context.WithCancel(context.Background())

	configuration, exitStatusCode := secrets.Main(ctx, cancel)

	removeTestDirectories(configuration)
	assert.Equal(t, contract.StatusCodeExitNormal, exitStatusCode)
}

func setupTest(args []string) func() {
	origArgs := os.Args
	os.Args = args
	fmt.Println("command line strings:", strings.Join(args, " "))

	origEnv, origEnvSet := os.LookupEnv(helper.EnvXdgRuntimeDir)
	if origEnvSet {
		os.Unsetenv(helper.EnvXdgRuntimeDir)
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	return func() {
		if origEnvSet {
			os.Setenv(helper.EnvXdgRuntimeDir, origEnv)
		} else {
			os.Unsetenv(helper.EnvXdgRuntimeDir)
		}
		os.Args = origArgs
		os.RemoveAll("./config")
	}
}

func removeTestDirectories(configuration *config.ConfigurationStruct) {
	remove := func(name string) {
		if name != "" && name != "/" {
			os.RemoveAll(name)
		}
	}
	if configuration != nil {
		remove(configuration.Logging.File)
		remove(configuration.SecretsSetup.WorkDir)
		remove(configuration.SecretsSetup.DeployDir)
		remove(configuration.SecretsSetup.CacheDir)
	}
}

func doesDirectoryExist(dir string) (bool, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false, err
	} else if err != nil {
		return true, err
	}
	return true, nil
}
