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

package help

import (
	"flag"
	"fmt"
	"strings"

	bootstrapper "github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/command"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

const (
	CommandName = "help"
)

type cmd struct {
	loggingClient logger.LoggingClient
	configuration *config.ConfigurationStruct
	flagSet       *flag.FlagSet
}

// NewCommand creates a new cmd and parses through options if any
func NewCommand(
	lc logger.LoggingClient,
	configuration *config.ConfigurationStruct,
	args []string) (interfaces.Command, error) {

	flagSet := flag.NewFlagSet(CommandName, flag.ContinueOnError)
	err := flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse command: %s: %w", strings.Join(args, " "), err)
	}

	return &cmd{
		loggingClient: lc,
		configuration: configuration,
		flagSet:       flagSet,
	}, nil
}

// Execute implements Command and runs this command
// command help prints the usage for security-bootstrapper
func (c *cmd) Execute() (statusCode int, err error) {
	bootstrapper.HelpCallback()
	return interfaces.StatusCodeExitNormal, nil
}

// GetCommandName returns the name of this command
func (c *cmd) GetCommandName() string {
	return CommandName
}
