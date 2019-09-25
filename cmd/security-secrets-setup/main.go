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

var subcommands = map[string]*flag.FlagSet{
	"legacy":   flag.NewFlagSet("legacy", flag.ExitOnError),
	"generate": flag.NewFlagSet("generate", flag.ExitOnError),
	"cache":    flag.NewFlagSet("cache", flag.ExitOnError),
}
var configFile string

func init() {
	// setup options for subcommands:
	subcommands["legacy"].StringVar(&configFile, "config", "", "specify JSON configuration file: /path/to/file.json")
	subcommands["legacy"].StringVar(&configFile, "c", "", "specify JSON configuration file: /path/to/file.json")

	flag.Usage = usage.HelpCallbackSecuritySetup
}

func main() {
	start := time.Now()

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Please specify subcommand for " + option.SecuritySecretsSetup)
		flag.Usage()
		exitInstance.exit(0)
		return
	}

	if err := setup.Init(); err != nil {
		// the error returned from Init has already been logged inside the call
		// so here we ignore the error logging
		exitInstance.exit(1)
		return
	}

	subcmdName := flag.Args()[0]

	subcmd, found := subcommands[subcmdName]
	if !found {
		setup.LoggingClient.Error(fmt.Sprintf("unsupported subcommand %s", subcmdName))
		exitInstance.exit(1)
		return
	}

	if err := subcmd.Parse(flag.Args()[1:]); err != nil {
		setup.LoggingClient.Error(fmt.Sprintf("error parsing subcommand %s: %v", subcmdName, err))
		exitInstance.exit(2)
		return
	}

	var exitStatusCode int
	var err error

	switch subcmdName {
	case "legacy":
		// no additional arguments expected
		if len(subcmd.Args()) > 0 {
			setup.LoggingClient.Error(fmt.Sprintf("subcommand %s doesn't use other additional args", subcmdName))
			exitInstance.exit(2)
			return
		}
		if err = option.GenTLSAssets(configFile); err != nil {
			setup.LoggingClient.Error(err.Error())
			exitInstance.exit(2)
			return
		}

	case "generate", "cache":
		// no arguments expected
		if len(subcmd.Args()) > 0 {
			setup.LoggingClient.Error(fmt.Sprintf("subcommand %s doesn't use any args", subcmdName))
			exitInstance.exit(2)
			return
		}
		exitStatusCode, err = dispatcherInstance.run(subcmdName)
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

func setupPkiInitOption(subcommand string) (executor option.OptionsExecutor, status int, err error) {
	var generateOpt, cacheOpt bool
	switch subcommand {
	case "generate":
		generateOpt = true
	case "cache":
		cacheOpt = true
	}

	opts := option.PkiInitOption{
		GenerateOpt: generateOpt,
		CacheOpt:    cacheOpt,
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
