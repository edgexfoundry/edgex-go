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

package option

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/edgexfoundry/edgex-go/internal/security/setup"
)

// Cache option....
func Cache() func(*PkiInitOption) (exitCode, error) {
	return func(pkiInitOpton *PkiInitOption) (exitCode, error) {

		if isCacheNoOp(pkiInitOpton) {
			return normal, nil
		}

		return cachePkis()
	}
}

func isCacheNoOp(pkiInitOption *PkiInitOption) bool {
	// nop: if the flag is missing or not on
	return pkiInitOption == nil || !pkiInitOption.CacheOpt
}

func cachePkis() (statusCode exitCode, err error) {
	// generate a new one if pkicache dir is empty
	pkiCacheDir := getPkiCacheDirEnv()
	cachedCAPrivateKeyFile := filepath.Join(pkiCacheDir, caServiceName, tlsSecretFileName)
	if empty, err := isDirEmpty(pkiCacheDir); err != nil {
		return exitWithError, err
	} else if empty {
		setup.LoggingClient.Info(fmt.Sprintf("cache dir %s is empty, generating PKI into it...", pkiCacheDir))
		// perform generate func
		statusCode, err = generatePkis()
		if err != nil {
			return statusCode, err
		}

		generatedDirPath := filepath.Join(os.Getenv(envXdgRuntimeDir), pkiInitGeneratedDir)

		// always shreds CA private key before cache
		caPrivateKeyFile := filepath.Join(generatedDirPath, caServiceName, tlsSecretFileName)
		if err := secureEraseFile(caPrivateKeyFile); err != nil {
			return exitWithError, err
		}

		if err = doCache(generatedDirPath); err != nil {
			return exitWithError, err
		}
	} else {
		// cache dir is not empty: output error message if CA private key is present
		// when cache is given
		if checkIfFileExists(cachedCAPrivateKeyFile) {
			return exitWithError, errCacheNotChangeAfter
		}
		setup.LoggingClient.Info(fmt.Sprintf("cached TLS assets from dir %s is present, using cached PKI", pkiCacheDir))
	}

	// to deploy
	// copy stuff into dest dir from pkiCache
	err = deploy(pkiCacheDir, pkiInitDeployDir)
	if err != nil {
		return exitWithError, err
	}

	return normal, nil
}

func doCache(fromDir string) error {
	// destination
	pkiCacheDir := getPkiCacheDirEnv()

	// to cache
	return copyDir(fromDir, pkiCacheDir)
}
