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
	"testing"

	option "github.com/edgexfoundry/security-secret-store/internal/pkg/pkiinit/option"
	"github.com/stretchr/testify/assert"
)

var hasDispatchError bool

func TestNoOption(t *testing.T) {
	tearDown := setupTest(t)
	origArgs := os.Args
	defer tearDown(t, origArgs)
	assert := assert.New(t)

	runWithNoOption()
	assert.Equal(1, (exitInstance.(*testExitCode)).getStatusCode())
	assert.Equal(false, helpOpt)
	assert.Equal(false, generateOpt)
}

func TestHelpOption(t *testing.T) {
	tearDown := setupTest(t)
	origArgs := os.Args
	defer tearDown(t, origArgs)
	assert := assert.New(t)

	runWithHelpOption()
	assert.Equal(0, (exitInstance.(*testExitCode)).getStatusCode())
	assert.Equal(true, helpOpt)
	assert.Equal(false, generateOpt)
}

func TestGenerateOptionOk(t *testing.T) {
	tearDown := setupTest(t)
	origArgs := os.Args
	defer tearDown(t, origArgs)
	assert := assert.New(t)

	runWithGenerateOption(false)
	assert.Equal(0, (exitInstance.(*testExitCode)).getStatusCode())
	assert.Equal(false, helpOpt)
	assert.Equal(true, generateOpt)
	optionExec := (dispatcherInstance.(*testPkiInitOptionDispatcher)).testOptsExecutor
	assert.Equal(true, (optionExec.(*option.PkiInitOption)).GenerateOpt)
}

func TestGenerateOptionWithRunError(t *testing.T) {
	tearDown := setupTest(t)
	origArgs := os.Args
	defer tearDown(t, origArgs)
	assert := assert.New(t)

	runWithGenerateOption(true)
	assert.Equal(2, (exitInstance.(*testExitCode)).getStatusCode())
	assert.Equal(false, helpOpt)
	assert.Equal(true, generateOpt)
}

func TestImportOptionOk(t *testing.T) {
	tearDown := setupTest(t)
	origArgs := os.Args
	defer tearDown(t, origArgs)
	assert := assert.New(t)

	runWithImportOption(false)
	assert.Equal(0, (exitInstance.(*testExitCode)).getStatusCode())
	assert.Equal(false, helpOpt)
	assert.Equal(true, importOpt)
	optionExec := (dispatcherInstance.(*testPkiInitOptionDispatcher)).testOptsExecutor
	assert.Equal(true, (optionExec.(*option.PkiInitOption)).ImportOpt)
}

func TestCacheOptionOk(t *testing.T) {
	tearDown := setupTest(t)
	origArgs := os.Args
	defer tearDown(t, origArgs)
	assert := assert.New(t)

	runWithCacheOption(false)
	assert.Equal(0, (exitInstance.(*testExitCode)).getStatusCode())
	assert.Equal(false, helpOpt)
	assert.Equal(true, cacheOpt)
	optionExec := (dispatcherInstance.(*testPkiInitOptionDispatcher)).testOptsExecutor
	assert.Equal(true, (optionExec.(*option.PkiInitOption)).CacheOpt)
}

func TestCacheCAOptionOk(t *testing.T) {
	tearDown := setupTest(t)
	origArgs := os.Args
	defer tearDown(t, origArgs)
	assert := assert.New(t)

	runWithCacheCAOption(false)
	assert.Equal(0, (exitInstance.(*testExitCode)).getStatusCode())
	assert.Equal(false, helpOpt)
	assert.Equal(true, cacheCAOpt)
	optionExec := (dispatcherInstance.(*testPkiInitOptionDispatcher)).testOptsExecutor
	assert.Equal(true, (optionExec.(*option.PkiInitOption)).CacheCAOpt)
}

func TestSetupPkiInitOption(t *testing.T) {
	tearDown := setupTest(t)
	origArgs := os.Args
	defer tearDown(t, origArgs)
	assert := assert.New(t)

	generateOpt = true

	_, status, err := setupPkiInitOption()
	assert.Equal(0, status)
	assert.Nil(err)
}

func setupTest(t *testing.T) func(t *testing.T, args []string) {
	exitInstance = newTestExit()
	dispatcherInstance = newTestDispatcher()
	return func(t *testing.T, args []string) {
		// reset after each test
		helpOpt = false
		generateOpt = false
		importOpt = false
		cacheOpt = false
		cacheCAOpt = false
		hasDispatchError = false
		os.Args = args
	}
}

func runWithNoOption() {
	// case 1: no option given
	os.Args = []string{"cmd"}
	printCommandLineStrings(os.Args)
	main()
}

func runWithHelpOption() {
	// case 2: h or help option given
	os.Args = []string{"cmd", "-help"}
	printCommandLineStrings(os.Args)
	main()
}

func runWithGenerateOption(hasError bool) {
	// case 3: generate option given
	os.Args = []string{"cmd", "-generate"}
	printCommandLineStrings(os.Args)
	hasDispatchError = hasError
	main()
}

func runWithImportOption(hasError bool) {
	// case 4: import option given
	os.Args = []string{"cmd", "-import"}
	printCommandLineStrings(os.Args)
	hasDispatchError = hasError
	main()
}

func runWithCacheOption(hasError bool) {
	// case 5: cache option given
	os.Args = []string{"cmd", "-cache"}
	printCommandLineStrings(os.Args)
	hasDispatchError = hasError
	main()
}

func runWithCacheCAOption(hasError bool) {
	// case 6: cache CA option given
	os.Args = []string{"cmd", "-cacheca"}
	printCommandLineStrings(os.Args)
	hasDispatchError = hasError
	main()
}

func printCommandLineStrings(strs []string) {
	fmt.Print("command line strings: ")
	for _, str := range strs {
		fmt.Print(str)
		fmt.Print(" ")
	}
	fmt.Println()
}

type testExitCode struct {
	testStatusCode int
}

func newTestExit() exit {
	return &testExitCode{}
}

func (testExit *testExitCode) callExit(statusCode int) {
	fmt.Printf("In test: exitCode = %d\n", statusCode)
	testExit.testStatusCode = statusCode
}

func (testExit *testExitCode) getStatusCode() int {
	return testExit.testStatusCode
}

type testPkiInitOptionDispatcher struct {
	testOptsExecutor option.OptionsExecutor
}

func newTestDispatcher() optionDispatcher {
	return &testPkiInitOptionDispatcher{}
}

func (testDispatcher *testPkiInitOptionDispatcher) run() (statusCode int, err error) {
	fmt.Printf("In test flag value: helpOpt = %v, generateOpt = %v, importOpt = %v, cacheOpt = %v, cacheCAOpt = %v \n",
		helpOpt, generateOpt, importOpt, cacheOpt, cacheCAOpt)

	optsExecutor, _, _ := setupPkiInitOption()
	testDispatcher.testOptsExecutor = optsExecutor

	if hasDispatchError {
		statusCode = 2
		err = errors.New("dispatch error found")
	}
	return statusCode, err
}
