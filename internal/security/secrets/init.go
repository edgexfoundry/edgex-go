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

	"github.com/edgexfoundry/edgex-go/internal/security/secrets/command/cache"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/command/generate"
	_import "github.com/edgexfoundry/edgex-go/internal/security/secrets/command/import"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/command/legacy"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/container"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/contract"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
)

type Bootstrap struct {
	exitStatusCode int
}

func NewBootstrapHandler() *Bootstrap {
	return &Bootstrap{}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the data service.
func (b *Bootstrap) Handler(
	ctx context.Context,
	wg *sync.WaitGroup,
	startupTimer startup.Timer,
	dic *di.Container) bool {

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)

	var command contract.Command
	var flagSet *flag.FlagSet

	commandName := flag.Args()[0]
	switch commandName {
	case legacy.CommandName:
		command, flagSet = legacy.NewCommand(lc)
	case generate.CommandName:
		command, flagSet = generate.NewCommand(lc, configuration)
	case cache.CommandName:
		generateCommand, _ := generate.NewCommand(lc, configuration)
		command, flagSet = cache.NewCommand(lc, configuration, generateCommand)
	case _import.CommandName:
		command, flagSet = _import.NewCommand(lc, configuration)
	default:
		lc.Error(fmt.Sprintf("unsupported subcommand %s", commandName))
		b.exitStatusCode = contract.StatusCodeNoOptionSelected
		return false
	}

	if err := flagSet.Parse(flag.Args()[1:]); err != nil {
		lc.Error(fmt.Sprintf("error parsing subcommand %s: %v", commandName, err))
		b.exitStatusCode = contract.StatusCodeExitWithError
		return false
	}

	if len(flagSet.Args()) > 0 {
		lc.Error(fmt.Sprintf("subcommand %s doesn't use any args", commandName))
		b.exitStatusCode = contract.StatusCodeExitWithError
		return false
	}

	exitStatusCode, err := command.Execute()
	if err != nil {
		lc.Error(err.Error())
	}
	b.exitStatusCode = exitStatusCode
	return false
}

func (b *Bootstrap) ExitStatusCode() int {
	return b.exitStatusCode
}
