/*******************************************************************************
 * Copyright 2025 IOTech Ltd
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
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/command/createtoken"
	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
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
// for security fileprovider's command arguments
func (b *Bootstrap) BootstrapHandler(_ context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	conf := container.ConfigurationFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	if flag.NArg() == 0 {
		lc.Debug("no security-file-provider subcommand arg defined, skipping execute subcommand")
		return true
	}

	lc.Debugf("configuration from the local file: %v", *conf)

	var command interfaces.Command
	var err error
	commandName := flag.Arg(0)
	subcommandArgs := flag.Args()

	lc.Debugf("security-file-provider commandName: %s, subcommandArgs: %v", commandName, subcommandArgs)

	switch commandName {
	case createtoken.CommandName:
		command, err = createtoken.NewCommand(dic, subcommandArgs[1:])
	default:
		return true
	}

	if err != nil {
		lc.Error(err.Error())
		b.exitStatusCode = interfaces.StatusCodeExitWithError
		return false
	}

	exitStatusCode, err := command.Execute()
	if err != nil {
		lc.Errorf("failed to execute command '%s': %v", commandName, err)
	}
	b.exitStatusCode = exitStatusCode

	return true
}

// GetExitStatusCode returns security fileprovider's exit code
func (b *Bootstrap) GetExitStatusCode() int {
	return b.exitStatusCode
}
