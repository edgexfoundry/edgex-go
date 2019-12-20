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

package legacy

import (
	"flag"

	"github.com/edgexfoundry/edgex-go/internal/security/secrets/contract"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/helper"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

const CommandName = "legacy"

type Command struct {
	flags         *flag.FlagSet
	configFile    string
	loggingClient logger.LoggingClient
}

func NewCommand(lc logger.LoggingClient) (*Command, *flag.FlagSet) {
	command := Command{
		loggingClient: lc,
	}
	flags := flag.NewFlagSet(CommandName, flag.ExitOnError)
	flags.StringVar(&command.configFile, "config", "", "specify JSON configuration file: /path/to/file.json")
	flags.StringVar(&command.configFile, "c", "", "specify JSON configuration file: /path/to/file.json")
	return &command, flags
}

func (c *Command) Execute() (statusCode int, err error) {
	err = helper.GenTLSAssets(c.configFile, c.loggingClient)
	if err != nil {
		statusCode = contract.StatusCodeExitWithError
	} else {
		statusCode = contract.StatusCodeExitNormal
	}
	return
}
