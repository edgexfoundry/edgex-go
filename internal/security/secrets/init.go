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
	"sync"

	bootstrapContainer "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/command/cache"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/command/generate"
	_import "github.com/edgexfoundry/edgex-go/internal/security/secrets/command/import"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/command/legacy"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/container"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/contract"
)

type Bootstrap struct {
	exitStatusCode int
}

func NewBootstrapHandler() *Bootstrap {
	return &Bootstrap{}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the data service.
func (b *Bootstrap) Handler(wg *sync.WaitGroup, ctx context.Context, startupTimer startup.Timer, dic *di.Container) bool {
	loggingClient := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)

	var command contract.Command
	var flagSet *flag.FlagSet

	commandName := flag.Args()[0]
	switch commandName {
	case legacy.CommandName:
		command, flagSet = legacy.NewCommand(loggingClient)
	case generate.CommandName:
		command, flagSet = generate.NewCommand(loggingClient, configuration)
	case cache.CommandName:
		generateCommand, _ := generate.NewCommand(loggingClient, configuration)
		command, flagSet = cache.NewCommand(loggingClient, configuration, generateCommand)
	case _import.CommandName:
		command, flagSet = _import.NewCommand(loggingClient, configuration)
	default:
		loggingClient.Error(fmt.Sprintf("unsupported subcommand %s", commandName))
		b.exitStatusCode = contract.StatusCodeNoOptionSelected
		return false
	}

	if err := flagSet.Parse(flag.Args()[1:]); err != nil {
		loggingClient.Error(fmt.Sprintf("error parsing subcommand %s: %v", commandName, err))
		b.exitStatusCode = contract.StatusCodeExitWithError
		return false
	}

	if len(flagSet.Args()) > 0 {
		loggingClient.Error(fmt.Sprintf("subcommand %s doesn't use any args", commandName))
		b.exitStatusCode = contract.StatusCodeExitWithError
		return false
	}

	exitStatusCode, err := command.Execute()
	if err != nil {
		loggingClient.Error(err.Error())
	}
	b.exitStatusCode = exitStatusCode
	return false
}

func (b *Bootstrap) ExitStatusCode() int {
	return b.exitStatusCode
}
