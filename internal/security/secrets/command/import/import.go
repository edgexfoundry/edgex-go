//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0'
//

package _import

import (
	"flag"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/security/secrets/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/contract"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/helper"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

const CommandName = "import"

type Command struct {
	loggingClient logger.LoggingClient
	configuration *config.ConfigurationStruct
}

func NewCommand(
	loggingClient logger.LoggingClient,
	configuration *config.ConfigurationStruct) (*Command, *flag.FlagSet) {

	return &Command{
		loggingClient: loggingClient,
		configuration: configuration,
	}, flag.NewFlagSet(CommandName, flag.ExitOnError)
}

// Execute deploys PKI from CacheDir to DeployDir.  It retruns an error,
// if CacheDir is empty instead of blocking or waiting;
// otherwise it just copies PKI assets from CacheDir into DeployDir.
// This import enables usage models for deploying a pre-populated PKI assets
// such as Kong TLS signed by an external certificate authority or TLS keys
// by other certificate authority.
func (c *Command) Execute() (statusCode int, err error) {
	pkiCacheDir, err := helper.GetCacheDir(c.configuration)
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}
	c.loggingClient.Info(fmt.Sprintf("importing from PKI cache dir: %s", pkiCacheDir))

	dirEmpty, err := helper.IsDirEmpty(pkiCacheDir)
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}
	if dirEmpty {
		return contract.StatusCodeExitWithError,
			fmt.Errorf("Expecting pre-populated PKI in the directory %s but found empty", pkiCacheDir)
	}

	// copy stuff into dest dir from pkiCache
	deployDir, err := helper.GetDeployDir(c.configuration)
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}
	err = helper.Deploy(pkiCacheDir, deployDir, c.loggingClient)
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}
	return statusCode, err
}
