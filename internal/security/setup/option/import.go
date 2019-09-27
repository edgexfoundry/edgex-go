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

	"github.com/edgexfoundry/edgex-go/internal/security/setup"
)

// Import deploys PKI from CacheDir to DeployDir.  It retruns an error,
// if CacheDir is empty instead of blocking or waiting;
// otherwise it just copies PKI assets from CacheDir into DeployDir.
// This import enables usage models for deploying a pre-populated PKI assets
// such as Kong TLS signed by an external certificate authority or TLS keys
// by other certificate authority.
func Import() func(*PkiInitOption) (exitCode, error) {
	return func(pkiInitOpton *PkiInitOption) (exitCode, error) {

		if isImportNoOp(pkiInitOpton) {
			return normal, nil
		}

		return importPkis()
	}
}

func isImportNoOp(pkiInitOption *PkiInitOption) bool {
	// nop: if the flag is missing or not on
	return pkiInitOption == nil || !pkiInitOption.ImportOpt
}

func importPkis() (statusCode exitCode, err error) {
	pkiCacheDir := getPkiCacheDirEnv()
	setup.LoggingClient.Info(fmt.Sprintf("importing from PKI_CACHE: %s", pkiCacheDir))

	dirEmpty, err := isDirEmpty(pkiCacheDir)

	if err != nil {
		return exitWithError, err
	}

	if !dirEmpty {
		// copy stuff into dest dir from pkiCache
		err = deploy(pkiCacheDir, pkiInitDeployDir)
		if err != nil {
			statusCode = exitWithError
		}
	} else {
		statusCode = exitWithError
		err = fmt.Errorf("Expecting pre-populated PKI in the directory %s but found empty", pkiCacheDir)
	}

	return statusCode, err
}
