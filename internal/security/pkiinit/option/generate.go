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
	"log"
	"os"
	"os/exec"
	"path/filepath"

	cert "github.com/edgexfoundry/edgex-go/internal/security/pkiinit/cert"
)

// Generate option....
func Generate() func(*PkiInitOption) (exitCode, error) {
	return func(pkiInitOpton *PkiInitOption) (exitCode, error) {

		if isGenerateNoOp(pkiInitOpton) {
			return normal, nil
		}

		if statusCode, err := generatePkis(); err != nil {
			return statusCode, err
		}
		generatedDirPath := filepath.Join(os.Getenv(envXdgRuntimeDir), pkiInitGeneratedDir)
		// Shred the CA private key before deploy
		caPrivateKeyFile := filepath.Join(generatedDirPath, caServiceName, tlsSecretFileName)
		if err := secureEraseFile(caPrivateKeyFile); err != nil {
			return exitWithError, err
		}

		if err := deploy(generatedDirPath, pkiInitDeployDir); err != nil {
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
	baseWorkingDir, err := os.Getwd()
	if err != nil {
		return exitWithError, err
	}

	pkiSetupRunPath := filepath.Join(baseWorkingDir, pkiSetupExecutable)
	pkiSetupVaultJSONPath := filepath.Join(baseWorkingDir, pkiSetupVaultJSON)

	scratchPath := filepath.Join(os.Getenv(envXdgRuntimeDir), pkiInitScratchDir)

	log.Println("pkiSetupRunPath: ", pkiSetupRunPath,
		"  pkiSetupVaultJSONPath: ", pkiSetupVaultJSONPath,
		"  scratchPath: ", scratchPath)

	if _, err := exec.LookPath(pkiSetupRunPath); err != nil {
		return exitWithError, err
	}

	if _, err := os.Stat(pkiSetupVaultJSONPath); os.IsNotExist(err) {
		log.Println("Vault JSON file for pkisetup not exists in ", pkiSetupVaultJSONPath)
		return exitWithError, err
	}

	// create scratch dir if not exists yet:
	if err := createDirectoryIfNotExists(scratchPath); err != nil {
		return exitWithError, err
	}

	// after done, need to change it back to the original working dir to avoid os.Getwd() error
	// and delete the scratch dir
	defer cleanup(baseWorkingDir, scratchPath)

	// run pkisetup binary on the env. of $XDG_RUNTIME_DIR/edgex/pki-init/scratch
	if err := os.Chdir(scratchPath); err != nil {
		return exitWithError, err
	}

	cmd := exec.Command(pkiSetupRunPath, "--config="+pkiSetupVaultJSONPath)
	result, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return exitWithError, execErr
	}

	log.Printf("result of pkisetup: %s\n", string(result))

	return rearrangePkiByServices(baseWorkingDir, pkiSetupVaultJSONPath)
}

func rearrangePkiByServices(baseWorkingDir, pkiSetupVaultJSONPath string) (exitCode, error) {
	reader := cert.NewX509ConfigReader(pkiSetupVaultJSONPath)

	if reader == nil {
		return exitWithError, fmt.Errorf("Failed to create X509ConfigReader with json path: %s", pkiSetupVaultJSONPath)
	}

	config, readErr := reader.Read()
	if readErr != nil {
		return exitWithError, readErr
	}

	pkiOutputDir, err := config.GetPkiOutputDir()
	if err != nil {
		return exitWithError, err
	}

	generatedDirPath := filepath.Join(os.Getenv(envXdgRuntimeDir), pkiInitGeneratedDir)

	log.Println("pki-init generate output base dir: ", generatedDirPath)

	// create generated dir if not exists yet:
	if err := createDirectoryIfNotExists(generatedDirPath); err != nil {
		return exitWithError, err
	}

	// CA:
	caDirPath := filepath.Join(generatedDirPath, caServiceName)
	err = copyGeneratedForService(caDirPath, pkiOutputDir, config)

	// Vault:
	vaultServicePath := filepath.Join(generatedDirPath, vaultServiceName)
	err = copyGeneratedForService(vaultServicePath, pkiOutputDir, config)

	return normal, nil
}

func copyGeneratedForService(servicePath, pkiOutputDir string, config cert.X509Config) error {
	if err := createDirectoryIfNotExists(servicePath); err != nil {
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

func cleanup(baseWorkingDir, scratchPath string) {
	os.Chdir(baseWorkingDir)
	os.RemoveAll(scratchPath)
	log.Println("pki-init generation completes")
}
