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
	"path/filepath"
)

const (
	// SecuritySecretsSetup is the name of this module
	SecuritySecretsSetup = "security-secrets-setup"

	pkiSetupVaultJSON             = "pkisetup-vault.json"
	pkiSetupKongJSON              = "pkisetup-kong.json"
	resourceDirName               = "res"
	configTomlFile                = "configuration.toml"
	envXdgRuntimeDir              = "XDG_RUNTIME_DIR"
	defaultXdgRuntimeDir          = "/tmp"
	pkiInitBaseDir                = "/edgex/security-secrets-setup"
	tmpfsRunDir                   = "/run"
	tlsSecretFileName             = "server.key"
	tlsCertFileName               = "server.crt"
	caCertFileName                = "ca.pem"
	pkiInitFilePerServiceComplete = ".security-secrets-setup.complete"

	// service name section:
	caServiceName    = "ca"
	vaultServiceName = "edgex-vault"
	kongServiceName  = "edgex-kong"
)

var (
	pkiInitScratchDir   = filepath.Join(pkiInitBaseDir, "scratch")
	pkiInitGeneratedDir = filepath.Join(pkiInitBaseDir, "generated")
	pkiInitDeployDir    = filepath.Join(tmpfsRunDir, "edgex", "secrets")
)
