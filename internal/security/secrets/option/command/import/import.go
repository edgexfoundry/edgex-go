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
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/command/constant"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

type Command struct {
	loggingClient logger.LoggingClient
}

func NewCommand(loggingClient logger.LoggingClient) *Command {
	return &Command{
		loggingClient: loggingClient,
	}
}

// Execute deploys PKI from CacheDir to DeployDir.  It retruns an error,
// if CacheDir is empty instead of blocking or waiting;
// otherwise it just copies PKI assets from CacheDir into DeployDir.
// This import enables usage models for deploying a pre-populated PKI assets
// such as Kong TLS signed by an external certificate authority or TLS keys
// by other certificate authority.
func (c *Command) Execute() (statusCode int, err error) {
	pkiCacheDir, err := option.GetCacheDir()
	if err != nil {
		return constant.ExitWithError, err
	}
	c.loggingClient.Info(fmt.Sprintf("importing from PKI cache dir: %s", pkiCacheDir))

	dirEmpty, err := option.IsDirEmpty(pkiCacheDir)

	if err != nil {
		return constant.ExitWithError, err
	}

	if !dirEmpty {
		// copy stuff into dest dir from pkiCache
		deployDir, err := option.GetDeployDir()
		if err != nil {
			return constant.ExitWithError, err
		}
		err = option.Deploy(pkiCacheDir, deployDir)
		if err != nil {
			statusCode = constant.ExitWithError
		}
	} else {
		statusCode = constant.ExitWithError
		err = fmt.Errorf("Expecting pre-populated PKI in the directory %s but found empty", pkiCacheDir)
	}

	return statusCode, err
}
