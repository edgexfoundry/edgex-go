/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2019 Intel Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/security/setup"
	"github.com/edgexfoundry/edgex-go/internal/security/setup/option"
)

type exiter interface {
	exit(int)
}

type exitCode struct{}

type optionDispatcher interface {
	run(command string) (int, error)
}

type pkiInitOptionDispatcher struct{}

var exitInstance = newExit()
var dispatcherInstance = newOptionDispatcher()
var configFile string

func init() {
	// define and register command line flags:
	flag.StringVar(&configFile, "config", "", "specify JSON configuration file: /path/to/file.json")
	flag.StringVar(&configFile, "c", "", "specify JSON configuration file: /path/to/file.json")

	flag.Usage = usage.HelpCallbackSecuritySetup
}

func main() {
	start := time.Now()

	if len(os.Args) < 2 {
		fmt.Println("Please specify subcommand or options for " + option.SecuritySecretsSetup)
		flag.Usage()
		exitInstance.exit(0)
		return
	}

	// Before we call Golang's flag.Parse(), we want to make sure that
	// the subcommand is extracted ans saved if subcommand is used
	// retrieve subcommand and delete it from os.Args[]
	// as the Golang flag.Parse() method always parses from
	// the frist arguments of os.Args[]
	subcommand := retrieveSubcommand()

	flag.Parse()

	if err := setup.Init(); err != nil {
		// the error returned from Init has already been logged inside the call
		// so here we ignore the error logging
		exitInstance.exit(1)
		return
	}

	if configFile == "" {
		// run with other options for pki-init
		statusCode, err := dispatcherInstance.run(subcommand)
		if err != nil {
			setup.LoggingClient.Error(err.Error())
		}

		exitInstance.exit(statusCode)
		return
	}

	if err := option.GenTLSAssets(configFile); err != nil {
		setup.LoggingClient.Error(err.Error())
		exitInstance.exit(2)
		return
	}

	setup.LoggingClient.Info(option.SecuritySecretsSetup+" complete", internal.LogDurationKey, time.Since(start).String())
}

func newExit() exiter {
	return &exitCode{}
}

func (code *exitCode) exit(statusCode int) {
	os.Exit(statusCode)
}

func newOptionDispatcher() optionDispatcher {
	return &pkiInitOptionDispatcher{}
}

// retrieveSubcommand parses the subcommand out of os.Arg if any
// otherwise returns empty string ("")
func retrieveSubcommand() (subcommand string) {
	// commandline options always starts with -
	// so it is treated as a subcommand if that is not the case
	if !strings.HasPrefix(os.Args[1], "-") {
		subcommand = os.Args[1]
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}
	return subcommand
}

func setupPkiInitOption(subcommand string) (executor option.OptionsExecutor, status int, err error) {
	generateOpt := false
	switch subcommand {
	case "generate":
		generateOpt = true
	default:
		return nil, 1, fmt.Errorf("unsupported subcommand %s", subcommand)
	}

	opts := option.PkiInitOption{
		GenerateOpt: generateOpt,
	}
	return option.NewPkiInitOption(opts)
}

func (dispatcher *pkiInitOptionDispatcher) run(command string) (statusCode int, err error) {
	optsExecutor, statusCode, err := setupPkiInitOption(command)
	if err != nil {
		return statusCode, err
	}

	return optsExecutor.ProcessOptions()
}
