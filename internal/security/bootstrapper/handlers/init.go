/*******************************************************************************
 * Copyright 2021 Intel Corporation
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
 *
 *******************************************************************************/

package handlers

import (
	"context"
	"flag"
	"os"
	"sync"

	bootstrapper "github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/command"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/command/help"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/container"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/interfaces"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

// Bootstrap is to implement BootstrapHandler
type Bootstrap struct {
	exitStatusCode int
}

// NewInitialization is to instantiate a Bootstrap instance
func NewInitialization() *Bootstrap {
	return &Bootstrap{}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed
// for security bootstrapper's command arguments
func (b *Bootstrap) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	conf := container.ConfigurationFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	lc.Debugf("configuration from the local TOML: %v", *conf)

	var command interfaces.Command
	var err error

	var confdir string
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flagSet.StringVar(&confdir, "confdir", "", "") // handled by bootstrap; duplicated here to prevent arg parsing errors
	err = flagSet.Parse(os.Args[1:])
	if err != nil {
		lc.Error(err.Error())
	}

	subcommandArgs := []string{}

	commandName := flagSet.Arg(0)
	if flag.NArg() > 0 {
		subcommandArgs = append(subcommandArgs, flag.Args()...)
	}

	switch commandName {
	case help.CommandName:
		command, err = help.NewCommand(lc, conf, subcommandArgs)
	default:
		command, err = bootstrapper.NewCommand(ctx, wg, lc, conf, subcommandArgs)
		if command == nil {
			lc.Error(err.Error())
			b.exitStatusCode = interfaces.StatusCodeNoOptionSelected
			return false
		}
	}

	if err != nil {
		lc.Error(err.Error())
		b.exitStatusCode = interfaces.StatusCodeExitWithError
		return false
	}

	exitStatusCode, err := command.Execute()
	if err != nil {
		lc.Error(err.Error())
	}
	b.exitStatusCode = exitStatusCode

	wg.Done()

	return true
}

// GetExitStatusCode returns security bootstrapper's exit code
func (b *Bootstrap) GetExitStatusCode() int {
	return b.exitStatusCode
}
