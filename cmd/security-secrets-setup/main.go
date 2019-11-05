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

	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option"
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
	"import":   flag.NewFlagSet("import", flag.ExitOnError),
}
var configFile string
var configDir string

func init() {
	// setup options for subcommands:
	subcommands["legacy"].StringVar(&configFile, "config", "", "specify JSON configuration file: /path/to/file.json")
	subcommands["legacy"].StringVar(&configFile, "c", "", "specify JSON configuration file: /path/to/file.json")

	flag.StringVar(&configDir, "confdir", "", "Specify local configuration directory")

	flag.Usage = usage.HelpCallbackSecuritySetup
}

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Please specify subcommand for " + option.SecuritySecretsSetup)
		flag.Usage()
		exitInstance.exit(0)
		return
	}

	if err := secrets.Init(configDir); err != nil {
		// the error returned from Init has already been logged inside the call
		// so here we ignore the error logging
		exitInstance.exit(1)
		return
	}

	subcmdName := flag.Args()[0]

	subcmd, found := subcommands[subcmdName]
	if !found {
		secrets.LoggingClient.Error(fmt.Sprintf("unsupported subcommand %s", subcmdName))
		exitInstance.exit(1)
		return
	}

	if err := subcmd.Parse(flag.Args()[1:]); err != nil {
		secrets.LoggingClient.Error(fmt.Sprintf("error parsing subcommand %s: %v", subcmdName, err))
		exitInstance.exit(2)
		return
	}

	var exitStatusCode int
	var err error

	switch subcmdName {
	case "legacy":
		// no additional arguments expected
		if len(subcmd.Args()) > 0 {
			secrets.LoggingClient.Error(fmt.Sprintf("subcommand %s doesn't use other additional args", subcmdName))
			exitInstance.exit(2)
			return
		}
		if err = option.GenTLSAssets(configFile); err != nil {
			secrets.LoggingClient.Error(err.Error())
			exitInstance.exit(2)
			return
		}

	case "generate", "cache", "import":
		// no arguments expected
		if len(subcmd.Args()) > 0 {
			secrets.LoggingClient.Error(fmt.Sprintf("subcommand %s doesn't use any args", subcmdName))
			exitInstance.exit(2)
			return
		}
		exitStatusCode, err = dispatcherInstance.run(subcmdName)
		if err != nil {
			secrets.LoggingClient.Error(err.Error())
		}
	}

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
	var generateOpt, cacheOpt, importOpt bool
	switch subcommand {
	case "generate":
		generateOpt = true
	case "cache":
		cacheOpt = true
	case "import":
		importOpt = true
	}

	opts := option.PkiInitOption{
		GenerateOpt: generateOpt,
		CacheOpt:    cacheOpt,
		ImportOpt:   importOpt,
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
