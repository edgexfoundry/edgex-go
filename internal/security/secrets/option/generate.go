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

	config "github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets"
)

// Generate option....
func Generate() func(*PkiInitOption) (exitCode, error) {
	return func(pkiInitOption *PkiInitOption) (exitCode, error) {

		if isGenerateNoOp(pkiInitOption) {
			return normal, nil
		}

		if statusCode, err := generatePkis(); err != nil {
			return statusCode, err
		}

		workDir, err := getWorkDir()
		if err != nil {
			return exitWithError, err
		}

		generatedDirPath := filepath.Join(workDir, pkiInitGeneratedDir)
		defer os.RemoveAll(generatedDirPath)

		// Shred the CA private key before deploy
		caPrivateKeyFile := filepath.Join(generatedDirPath, caServiceName, tlsSecretFileName)
		if err := secureEraseFile(caPrivateKeyFile); err != nil {
			return exitWithError, err
		}

		deployDir, err := getDeployDir()
		if err != nil {
			return exitWithError, err
		}

		if err := deploy(generatedDirPath, deployDir); err != nil {
			return exitWithError, err
		}

		return normal, nil
	}
}

func isGenerateNoOp(pkiInitOption *PkiInitOption) bool {
	// nop: if the flag is missing or not on
	return pkiInitOption == nil || !pkiInitOption.GenerateOpt
}

func generatePkis() (exitCode, error) {
	certConfigDir, err := getCertConfigDir()
	if err != nil {
		return exitWithError, err
	}

	certConfigDir, err = filepath.Abs(certConfigDir)
	if err != nil {
		return exitWithError, err
	}
	pkiSetupVaultJSONPath := filepath.Join(certConfigDir, pkiSetupVaultJSON)
	pkiSetupKongJSONPath := filepath.Join(certConfigDir, pkiSetupKongJSON)

	workingDir, err := getWorkDir()
	if err != nil {
		return exitWithError, err
	}
	scratchPath := filepath.Join(workingDir, pkiInitScratchDir)

	secrets.LoggingClient.Debug(fmt.Sprint("pkiSetupVaultJSONPath: ", pkiSetupVaultJSONPath,
		"  pkiSetupKongJSONPath: ", pkiSetupKongJSONPath,
		"  scratchPath: ", scratchPath,
		"  certConfigDir: ", certConfigDir))

	if !checkIfFileExists(pkiSetupVaultJSONPath) {
		return exitWithError, fmt.Errorf("Vault JSON file for security-secrets-setup does not exist in %s", pkiSetupVaultJSONPath)
	}

	if !checkIfFileExists(pkiSetupKongJSONPath) {
		return exitWithError, fmt.Errorf("Kong JSON file for security-secrets-setup does not exist in %s", pkiSetupKongJSONPath)
	}

	// create scratch dir if not exists yet:
	if err := createDirectoryIfNotExists(scratchPath); err != nil {
		return exitWithError, err
	}

	currDir, err := os.Getwd()
	if err != nil {
		return exitWithError, err
	}

	// after done, need to change it back to the original working dir to avoid os.Getwd() error
	// and delete the scratch dir
	defer cleanup(currDir, scratchPath)

	// generate TLS certs on the env. of $XDG_RUNTIME_DIR/edgex/pki-init/scratch
	if err := os.Chdir(scratchPath); err != nil {
		return exitWithError, err
	}

	if err := GenTLSAssets(pkiSetupVaultJSONPath); err != nil {
		return exitWithError, err
	}

	if err := GenTLSAssets(pkiSetupKongJSONPath); err != nil {
		return exitWithError, err
	}

	return rearrangePkiByServices(workingDir, pkiSetupVaultJSONPath, pkiSetupKongJSONPath)
}

func rearrangePkiByServices(workingDir, pkiSetupVaultJSONPath, pkiSetupKongJSONPath string) (exitCode, error) {
	vaultConfig, readErr := config.NewX509Config(pkiSetupVaultJSONPath)
	if readErr != nil {
		return exitWithError, readErr
	}

	kongConfig, readErr := config.NewX509Config(pkiSetupKongJSONPath)
	if readErr != nil {
		return exitWithError, readErr
	}

	generatedDirPath := filepath.Join(workingDir, pkiInitGeneratedDir)

	secrets.LoggingClient.Debug(fmt.Sprint("pki-init generate output base dir: ", generatedDirPath))

	// create generated dir if not exists yet:
	if err := createDirectoryIfNotExists(generatedDirPath); err != nil {
		return exitWithError, err
	}

	// CA:
	caDirPath := filepath.Join(generatedDirPath, caServiceName)
	if err := copyGeneratedForService(caDirPath, vaultConfig); err != nil {
		return exitWithError, err
	}

	// Vault:
	vaultServicePath := filepath.Join(generatedDirPath, vaultServiceName)
	if err := copyGeneratedForService(vaultServicePath, vaultConfig); err != nil {
		return exitWithError, err
	}

	// Kong:
	kongServicePath := filepath.Join(generatedDirPath, kongServiceName)
	if err := copyGeneratedForService(kongServicePath, kongConfig); err != nil {
		return exitWithError, err
	}

	return normal, nil
}

func copyGeneratedForService(servicePath string, config config.X509Config) error {
	if err := createDirectoryIfNotExists(servicePath); err != nil {
		return err
	}

	pkiOutputDir, err := config.PkiCADir()
	if err != nil {
		return err
	}

	if _, err := copyFile(filepath.Join(pkiOutputDir, config.GetCAPemFileName()), filepath.Join(servicePath, caCertFileName)); err != nil {
		return err
	}

	privKeyFileName := filepath.Join(servicePath, tlsSecretFileName)
	if filepath.Base(servicePath) == caServiceName {
		if _, err := copyFile(filepath.Join(pkiOutputDir, config.GetCAPrivateKeyFileName()), privKeyFileName); err != nil {
			return err
		}
	} else {
		if _, err := copyFile(filepath.Join(pkiOutputDir, config.GetTLSPrivateKeyFileName()), privKeyFileName); err != nil {
			return err
		}
		// if not CA, then also copy the TLS cert as well
		if _, err := copyFile(filepath.Join(pkiOutputDir, config.GetTLSPemFileName()), filepath.Join(servicePath, tlsCertFileName)); err != nil {
			return err
		}
	}

	// read-only to the owner
	return os.Chmod(privKeyFileName, 0400)
}

func cleanup(origWorkingDir, scratchPath string) {
	_ = os.Chdir(origWorkingDir)
	os.RemoveAll(scratchPath)
	secrets.LoggingClient.Info("pki-init generation completes")
}
