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
var subcommandList = []string{"legacy", "generate", "cache", "import"}

func main() {
	start := time.Now()

	// define and register command line subcommands:
	legacyCmd := flag.NewFlagSet("legacy", flag.ExitOnError)
	legacyCmd.StringVar(&configFile, "config", "", "specify JSON configuration file: /path/to/file.json")
	legacyCmd.StringVar(&configFile, "c", "", "specify JSON configuration file: /path/to/file.json")

	flag.Usage = usage.HelpCallbackSecuritySetup

	flag.Parse()

	if len(os.Args) < 2 || flag.NArg() < 1 {
		fmt.Println("Please specify subcommand for " + option.SecuritySecretsSetup)
		flag.Usage()
		exitInstance.exit(0)
		return
	}

	subcommand := flag.Args()[0]

	if err := setup.Init(); err != nil {
		// the error returned from Init has already been logged inside the call
		// so here we ignore the error logging
		exitInstance.exit(1)
		return
	}

	if checkIfMultipleSubcommands() {
		setup.LoggingClient.Error("cannot use multiple subcommands, use one at a time")
		exitInstance.exit(1)
		return
	}

	setup.LoggingClient.Debug(fmt.Sprintf("subcommand <%s>", subcommand))

	var exitStatusCode int
	var err error
	// legacy mode of operation
	if "legacy" == subcommand {
		err = legacyCmd.Parse(flag.Args()[1:])
		if err != nil {
			setup.LoggingClient.Error(err.Error())
			exitInstance.exit(2)
			return
		}
		if err = option.GenTLSAssets(configFile); err != nil {
			setup.LoggingClient.Error(err.Error())
			exitInstance.exit(2)
			return
		}
	} else {
		// pki-init mode of operations
		exitStatusCode, err = dispatcherInstance.run(subcommand)
		if err != nil {
			setup.LoggingClient.Error(err.Error())
		}
	}

	setup.LoggingClient.Info(option.SecuritySecretsSetup+" complete", internal.LogDurationKey, time.Since(start).String())
	exitInstance.exit(exitStatusCode)
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

func checkIfMultipleSubcommands() bool {
	numSubCmds := 0
	for _, arg := range flag.Args() {
		for _, subcommand := range subcommandList {
			if arg == subcommand {
				numSubCmds++
			}
		}
	}
	return numSubCmds > 1
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
