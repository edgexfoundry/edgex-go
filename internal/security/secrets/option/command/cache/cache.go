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
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/command/generate"
	"os"
	"path/filepath"

	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

type Command struct {
	loggingClient logger.LoggingClient
	generate      *generate.Command
}

func NewCommand(loggingClient logger.LoggingClient, generate *generate.Command) *Command {
	return &Command{
		loggingClient: loggingClient,
		generate:      generate,
	}
}

// Execute generates PKI exactly once and cached to a designated location for future use.
// The PKI is then deployed from the cached location.
func (c *Command) Execute() (statusCode option.ExitCode, err error) {
	// generate a new one if pkicache dir is empty
	pkiCacheDir, err := option.GetCacheDir()
	if err != nil {
		return option.ExitWithError, err
	}

	empty, err := option.IsDirEmpty(pkiCacheDir)
	if err != nil {
		return option.ExitWithError, err
	}

	if empty {
		c.loggingClient.Info(fmt.Sprintf("cache dir %s is empty, generating PKI into it...", pkiCacheDir))
		// perform generate func

		if statusCode, err = c.generate.GeneratePkis(); err != nil {
			return statusCode, err
		}

		workDir, err := option.GetWorkDir()
		if err != nil {
			return option.ExitWithError, err
		}
		generatedDirPath := filepath.Join(workDir, option.PkiInitGeneratedDir)
		defer os.RemoveAll(generatedDirPath)

		// always shreds CA private key before cache
		caPrivateKeyFile := filepath.Join(generatedDirPath, option.CaServiceName, option.TlsSecretFileName)
		if err := option.SecureEraseFile(caPrivateKeyFile); err != nil {
			return option.ExitWithError, err
		}

		if err = c.doCache(generatedDirPath); err != nil {
			return option.ExitWithError, err
		}
	} else {
		// cache dir is not empty: output error message if CA private key is present
		// when cache is given
		cachedCAPrivateKeyFile := filepath.Join(pkiCacheDir, option.CaServiceName, option.TlsSecretFileName)
		if option.CheckIfFileExists(cachedCAPrivateKeyFile) {
			return option.ExitWithError, option.ErrCacheNotChangeAfter
		}
		c.loggingClient.Info(fmt.Sprintf("cached TLS assets from dir %s is present, using cached PKI", pkiCacheDir))
	}

	// to Deploy
	// copy stuff into dest dir from pkiCache
	deployDir, err := option.GetDeployDir()
	if err != nil {
		return option.ExitWithError, err
	}

	err = option.Deploy(pkiCacheDir, deployDir)
	if err != nil {
		return option.ExitWithError, err
	}

	return option.Normal, nil
}

func (c *Command) doCache(fromDir string) error {
	// destination
	pkiCacheDir, err := option.GetCacheDir()
	if err != nil {
		return err
	}

	// to cache
	return option.CopyDir(fromDir, pkiCacheDir)
}
