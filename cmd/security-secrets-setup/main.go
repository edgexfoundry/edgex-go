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
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/command/cache"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/command/generate"
	_import "github.com/edgexfoundry/edgex-go/internal/security/secrets/option/command/import"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/command/legacy"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/contract"
	"os"

	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/constant"
)

func main() {
	var configDir string

	legacyFlags := legacy.NewFlags()
	generateFlagSet := generate.NewFlags()
	cacheFlagSet := cache.NewFlags()
	importFlagSet := _import.NewFlags()
	flag.StringVar(&configDir, "confdir", "", "Specify local configuration directory")
	flag.Usage = usage.HelpCallbackSecuritySetup
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Please specify subcommand for " + constant.SecuritySecretsSetup)
		flag.Usage()
		os.Exit(constant.ExitNormal)
	}

	if err := secrets.Init(configDir); err != nil {
		// the error returned from Init has already been logged inside the call
		// so here we ignore the error logging
		os.Exit(constant.NoOptionSelected)
	}

	commandName := flag.Args()[0]
	var command contract.Command
	var flagSet *flag.FlagSet
	switch commandName {
	case constant.CommandLegacy:
		command, flagSet = legacy.NewCommand(legacyFlags)
	case constant.CommandGenerate:
		command, flagSet = generate.NewCommand(generateFlagSet, secrets.LoggingClient)
	case constant.CommandCache:
		generateCommand, _ := generate.NewCommand(generateFlagSet, secrets.LoggingClient)
		command, flagSet = cache.NewCommand(cacheFlagSet, secrets.LoggingClient, generateCommand)
	case constant.CommandImport:
		command, flagSet = _import.NewCommand(importFlagSet, secrets.LoggingClient)
	default:
		secrets.LoggingClient.Error(fmt.Sprintf("unsupported subcommand %s", commandName))
		os.Exit(constant.ExitWithError)
	}

	if err := flagSet.Parse(flag.Args()[1:]); err != nil {
		secrets.LoggingClient.Error(fmt.Sprintf("error parsing subcommand %s: %v", commandName, err))
		os.Exit(constant.ExitWithError)
	}

	if len(flagSet.Args()) > 0 {
		secrets.LoggingClient.Error(fmt.Sprintf("subcommand %s doesn't use any args", commandName))
		os.Exit(constant.ExitWithError)
	}

	exitStatusCode, err := command.Execute()
	if err != nil {
		secrets.LoggingClient.Error(err.Error())
	}
	os.Exit(exitStatusCode)
}
