/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package secrets

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"

	bootstrapContainer "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/command/cache"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/command/generate"
	_import "github.com/edgexfoundry/edgex-go/internal/security/secrets/option/command/import"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/command/legacy"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/contract"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// Global variables
var Configuration *config.ConfigurationStruct
var LoggingClient logger.LoggingClient

type Bootstrap struct {
	legacyFlags     *legacy.FlagSet
	generateFlagSet *generate.FlagSet
	cacheFlagSet    *cache.FlagSet
	importFlagSet   *_import.FlagSet
	commandName     string
}

func NewBootstrapHandler(
	legacyFlags *legacy.FlagSet,
	generateFlagSet *generate.FlagSet,
	cacheFlagSet *cache.FlagSet,
	importFlagSet *_import.FlagSet,
	commandName string) *Bootstrap {

	return &Bootstrap{
		legacyFlags:     legacyFlags,
		generateFlagSet: generateFlagSet,
		cacheFlagSet:    cacheFlagSet,
		importFlagSet:   importFlagSet,
		commandName:     commandName,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the data service.
func (b *Bootstrap) Handler(wg *sync.WaitGroup, ctx context.Context, startupTimer startup.Timer, dic *di.Container) bool {
	loggingClient := bootstrapContainer.LoggingClientFrom(dic.Get)

	commandName := flag.Args()[0]
	var command contract.Command
	var flagSet *flag.FlagSet
	switch commandName {
	case legacy.CommandLegacy:
		command, flagSet = legacy.NewCommand(b.legacyFlags)
	case generate.CommandGenerate:
		command, flagSet = generate.NewCommand(b.generateFlagSet, loggingClient)
	case cache.CommandCache:
		generateCommand, _ := generate.NewCommand(b.generateFlagSet, loggingClient)
		command, flagSet = cache.NewCommand(b.cacheFlagSet, loggingClient, generateCommand)
	case _import.CommandImport:
		command, flagSet = _import.NewCommand(b.importFlagSet, loggingClient)
	default:
		loggingClient.Error(fmt.Sprintf("unsupported subcommand %s", commandName))
		os.Exit(contract.StatusCodeExitWithError)
	}

	if err := flagSet.Parse(flag.Args()[1:]); err != nil {
		loggingClient.Error(fmt.Sprintf("error parsing subcommand %s: %v", commandName, err))
		os.Exit(contract.StatusCodeExitWithError)
	}

	if len(flagSet.Args()) > 0 {
		loggingClient.Error(fmt.Sprintf("subcommand %s doesn't use any args", commandName))
		os.Exit(contract.StatusCodeExitWithError)
	}

	exitStatusCode, err := command.Execute()
	if err != nil {
		loggingClient.Error(err.Error())
	}
	os.Exit(exitStatusCode)
}
