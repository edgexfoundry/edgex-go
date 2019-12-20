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

package cache

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/edgexfoundry/edgex-go/internal/security/secrets/command/generate"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/contract"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/helper"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

const CommandName = "cache"

type Command struct {
	loggingClient logger.LoggingClient
	configuration *config.ConfigurationStruct
	generate      *generate.Command
}

func NewCommand(
	lc logger.LoggingClient,
	configuration *config.ConfigurationStruct,
	generate *generate.Command) (*Command, *flag.FlagSet) {

	return &Command{
		loggingClient: lc,
		configuration: configuration,
		generate:      generate,
	}, flag.NewFlagSet(CommandName, flag.ExitOnError)
}

// Execute generates PKI exactly once and cached to a designated location for future use.
// The PKI is then deployed from the cached location.
func (c *Command) Execute() (statusCode int, err error) {
	// generate a new one if pkicache dir is empty
	pkiCacheDir, err := helper.GetCacheDir(c.configuration)
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}

	empty, err := helper.IsDirEmpty(pkiCacheDir)
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}

	if empty {
		c.loggingClient.Info(fmt.Sprintf("cache dir %s is empty, generating PKI into it...", pkiCacheDir))
		// perform generate func

		if statusCode, err = c.generate.GeneratePkis(); err != nil {
			return statusCode, err
		}

		workDir, err := helper.GetWorkDir(c.configuration)
		if err != nil {
			return contract.StatusCodeExitWithError, err
		}
		generatedDirPath := filepath.Join(workDir, generate.PkiInitGeneratedDir)
		defer os.RemoveAll(generatedDirPath)

		// always shreds CA private key before cache
		caPrivateKeyFile := filepath.Join(generatedDirPath, generate.CaServiceName, generate.TlsSecretFileName)
		if err := helper.SecureEraseFile(caPrivateKeyFile); err != nil {
			return contract.StatusCodeExitWithError, err
		}

		if err = c.doCache(generatedDirPath); err != nil {
			return contract.StatusCodeExitWithError, err
		}
	} else {
		// cache dir is not empty: output error message if CA private key is present
		// when cache is given
		cachedCAPrivateKeyFile := filepath.Join(pkiCacheDir, generate.CaServiceName, generate.TlsSecretFileName)
		if helper.CheckIfFileExists(cachedCAPrivateKeyFile) {
			return contract.StatusCodeExitWithError, errors.New("PKI cache cannot be changed after it was cached previously")
		}
		c.loggingClient.Info(fmt.Sprintf("cached TLS assets from dir %s is present, using cached PKI", pkiCacheDir))
	}

	// to Deploy
	// copy stuff into dest dir from pkiCache
	deployDir, err := helper.GetDeployDir(c.configuration)
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}

	err = helper.Deploy(pkiCacheDir, deployDir, c.loggingClient)
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}

	return contract.StatusCodeExitNormal, nil
}

func (c *Command) doCache(fromDir string) error {
	// destination
	pkiCacheDir, err := helper.GetCacheDir(c.configuration)
	if err != nil {
		return err
	}

	// to cache
	return helper.CopyDir(fromDir, pkiCacheDir, c.loggingClient)
}
