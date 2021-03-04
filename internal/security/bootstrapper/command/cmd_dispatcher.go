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

package command

import (
	"context"
	"fmt"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/command/gate"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/command/genpassword"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/command/gethttpstatus"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/command/listen"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/command/ping"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/command/setupacl"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/command/waitfor"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

// NewCommand instantiates a command implementing interfaces.Command based on the input command argument
func NewCommand(
	ctx context.Context,
	wg *sync.WaitGroup,
	lc logger.LoggingClient,
	configuration *config.ConfigurationStruct,
	args []string) (interfaces.Command, error) {

	var command interfaces.Command
	var err error

	if len(args) < 1 {
		return nil, fmt.Errorf("subcommand required (%s, %s, %s, %s, %s, %s, %s)", gate.CommandName, listen.CommandName,
			ping.CommandName, gethttpstatus.CommandName, genpassword.CommandName, waitfor.CommandName, setupacl.CommandName)
	}

	commandName := args[0]

	switch commandName {
	case gate.CommandName:
		command, err = gate.NewCommand(ctx, wg, lc, configuration, args[1:])
	case listen.CommandName:
		command, err = listen.NewCommand(ctx, wg, lc, configuration, args[1:])
	case ping.CommandName:
		command, err = ping.NewCommand(ctx, wg, lc, configuration, args[1:])
	case gethttpstatus.CommandName:
		command, err = gethttpstatus.NewCommand(ctx, wg, lc, configuration, args[1:])
	case genpassword.CommandName:
		command, err = genpassword.NewCommand(ctx, wg, lc, configuration, args[1:])
	case waitfor.CommandName:
		command, err = waitfor.NewCommand(ctx, wg, lc, configuration, args[1:])
	case setupacl.CommandName:
		command, err = setupacl.NewCommand(ctx, wg, lc, configuration, args[1:])
	default:
		command = nil
		err = fmt.Errorf("unsupported command %s", commandName)
	}

	return command, err
}
