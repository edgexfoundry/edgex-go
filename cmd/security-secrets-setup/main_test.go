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
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/setup/option"
	"github.com/stretchr/testify/assert"
)

type testExitCode struct {
	testStatusCode int
}

type testPkiInitOptionDispatcher struct {
	testOptsExecutor option.OptionsExecutor
}

var hasDispatchError bool

func TestMainWithNoOption(t *testing.T) {
	tearDown := setupTest(t)
	origArgs := os.Args
	defer tearDown(t, origArgs)
	assert := assert.New(t)

	os.Args = []string{"cmd"}
	printCommandLineStrings(os.Args)
	main()

	assert.Equal(0, (exitInstance.(*testExitCode)).getStatusCode())
}

func TestMainWithConfigFileOption(t *testing.T) {
	tearDown := setupTest(t)
	origArgs := os.Args
	defer tearDown(t, origArgs)
	assert := assert.New(t)

	os.Args = []string{"cmd", "legacy", "-config", "./res/pkisetup-vault.json"}
	printCommandLineStrings(os.Args)
	main()

	assert.Equal(0, (exitInstance.(*testExitCode)).getStatusCode())

	// verify ./config/pki/EdgeXFoundryCA directory exists
	exists, err := doesDirectoryExist("./config/pki/EdgeXFoundryCA")
	if !exists {
		assert.NotNil(err)
		assert.FailNow("cannot find the directory for TLS assets", err)
	}
}

func TestConfigFileOptionError(t *testing.T) {
	tearDown := setupTest(t)
	origArgs := os.Args
	defer tearDown(t, origArgs)
	assert := assert.New(t)

	os.Args = []string{"cmd", "legacy", "-config", "./non-exist/cert.json"}
	printCommandLineStrings(os.Args)
	main()

	assert.Equal(2, (exitInstance.(*testExitCode)).getStatusCode())
}

func TestMainWithGenerateOption(t *testing.T) {
	tearDown := setupTest(t)
	origArgs := os.Args
	defer tearDown(t, origArgs)
	assert := assert.New(t)

	runWithGenerateOption(false)
	assert.Equal(0, (exitInstance.(*testExitCode)).getStatusCode())
	optionExec := (dispatcherInstance.(*testPkiInitOptionDispatcher)).testOptsExecutor
	assert.True((optionExec.(*option.PkiInitOption)).GenerateOpt)
}

func TestGenerateOptionWithRunError(t *testing.T) {
	tearDown := setupTest(t)
	origArgs := os.Args
	defer tearDown(t, origArgs)
	assert := assert.New(t)

	runWithGenerateOption(true)
	assert.Equal(2, (exitInstance.(*testExitCode)).getStatusCode())
}

func TestMainUnsupportedArgument(t *testing.T) {
	tearDown := setupTest(t)
	origArgs := os.Args
	defer tearDown(t, origArgs)
	assert := assert.New(t)

	os.Args = []string{"cmd", "unsupported"}
	printCommandLineStrings(os.Args)
	hasDispatchError = false
	main()

	assert.Equal(1, (exitInstance.(*testExitCode)).getStatusCode())
}

func TestMainVerifyMultipleSubcommands(t *testing.T) {
	tearDown := setupTest(t)
	origArgs := os.Args
	defer tearDown(t, origArgs)
	assert := assert.New(t)

	os.Args = []string{"cmd", "generate", "legacy"}
	printCommandLineStrings(os.Args)
	hasDispatchError = false
	main()

	assert.Equal(2, (exitInstance.(*testExitCode)).getStatusCode())
}

func TestMainLegacySubcommandWithExtraArgs(t *testing.T) {
	tearDown := setupTest(t)
	origArgs := os.Args
	defer tearDown(t, origArgs)
	assert := assert.New(t)

	os.Args = []string{"cmd", "legacy", "-c", "./res/pkisetup-vault.json", "extra"}
	printCommandLineStrings(os.Args)
	hasDispatchError = false
	main()

	assert.Equal(2, (exitInstance.(*testExitCode)).getStatusCode())
}

func runWithGenerateOption(hasError bool) {
	os.Args = []string{"cmd", "generate"}
	printCommandLineStrings(os.Args)
	hasDispatchError = hasError
	main()
}

func printCommandLineStrings(strs []string) {
	fmt.Println("command line strings:", strings.Join(strs, " "))
}

func setupTest(t *testing.T) func(t *testing.T, args []string) {
	exitInstance = newTestExit()
	dispatcherInstance = newTestDispatcher()
	return func(t *testing.T, args []string) {
		// reset after each test
		configFile = ""
		os.Args = args
		// cleanup the generated files
		os.RemoveAll("./config")
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

func newTestExit() exiter {
	return &testExitCode{}
}

func (testExit *testExitCode) exit(statusCode int) {
	fmt.Printf("In test: exitCode = %d\n", statusCode)
	testExit.testStatusCode = statusCode
}

func (testExit *testExitCode) getStatusCode() int {
	return testExit.testStatusCode
}

func newTestDispatcher() optionDispatcher {
	return &testPkiInitOptionDispatcher{}
}

func (testDispatcher *testPkiInitOptionDispatcher) run(command string) (statusCode int, err error) {
	optsExecutor, statusCode, err := setupPkiInitOption(command)

	genOpt := false
	if pkiInit, ok := optsExecutor.(*option.PkiInitOption); ok {
		genOpt = pkiInit.GenerateOpt
	}

	fmt.Printf("In test flag value: generateOpt = %v, configFile = %v", genOpt, configFile)

	testDispatcher.testOptsExecutor = optsExecutor

	if hasDispatchError {
		statusCode = 2
		err = errors.New("dispatch error found")
	}
	return statusCode, err
}
