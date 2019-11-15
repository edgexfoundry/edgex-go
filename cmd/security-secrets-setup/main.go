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
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/command/cache"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/command/constant"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/command/generate"
	_import "github.com/edgexfoundry/edgex-go/internal/security/secrets/option/command/import"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/command/legacy"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/contract"
)

func main() {
	var configFile string
	var configDir string
	var subcommands = map[string]*flag.FlagSet{
		constant.CommandLegacy:   flag.NewFlagSet(constant.CommandLegacy, flag.ExitOnError),
		constant.CommandGenerate: flag.NewFlagSet(constant.CommandGenerate, flag.ExitOnError),
		constant.CommandCache:    flag.NewFlagSet(constant.CommandCache, flag.ExitOnError),
		constant.CommandImport:   flag.NewFlagSet(constant.CommandImport, flag.ExitOnError),
	}

	// setup options for subcommands:
	subcommands[constant.CommandLegacy].StringVar(&configFile, "config", "", "specify JSON configuration file: /path/to/file.json")
	subcommands[constant.CommandLegacy].StringVar(&configFile, "c", "", "specify JSON configuration file: /path/to/file.json")

	flag.StringVar(&configDir, "confdir", "", "Specify local configuration directory")

	flag.Usage = usage.HelpCallbackSecuritySetup
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Please specify subcommand for " + constant.SecuritySecretsSetup)
		flag.Usage()
		os.Exit(constant.ExitNormal)
		return
	}

	if err := secrets.Init(configDir); err != nil {
		// the error returned from Init has already been logged inside the call
		// so here we ignore the error logging
		os.Exit(constant.NoOptionSelected)
		return
	}

	subcmdName := flag.Args()[0]

	subcmd, found := subcommands[subcmdName]
	if !found {
		secrets.LoggingClient.Error(fmt.Sprintf("unsupported subcommand %s", subcmdName))
		os.Exit(constant.NoOptionSelected)
		return
	}

	if err := subcmd.Parse(flag.Args()[1:]); err != nil {
		secrets.LoggingClient.Error(fmt.Sprintf("error parsing subcommand %s: %v", subcmdName, err))
		os.Exit(constant.ExitWithError)
		return
	}

	// no arguments expected
	if len(subcmd.Args()) > 0 {
		secrets.LoggingClient.Error(fmt.Sprintf("subcommand %s doesn't use any args", subcmdName))
		os.Exit(constant.ExitWithError)
		return
	}

	var command contract.Command
	switch subcmdName {
	case constant.CommandLegacy:
		command = legacy.NewCommand(configFile)
	case constant.CommandGenerate:
		command = generate.NewCommand(secrets.LoggingClient)
	case constant.CommandCache:
		command = cache.NewCommand(secrets.LoggingClient, generate.NewCommand(secrets.LoggingClient))
	case constant.CommandImport:
		command = _import.NewCommand(secrets.LoggingClient)
	default:
		panic("unexpected subcmdName")
	}

	exitStatusCode, err := command.Execute()
	if err != nil {
		secrets.LoggingClient.Error(err.Error())
	}
	os.Exit(exitStatusCode)
}
