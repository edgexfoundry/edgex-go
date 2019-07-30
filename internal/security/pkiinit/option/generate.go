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
	"log"
	"os"
	"os/exec"
	"path/filepath"

	config "github.com/edgexfoundry/edgex-go/internal/pkg/config"
)

type pkiSetupRunnable interface {
	setRunConfig(pkiSetupRunConfig)
	checkRunConfig() error
	call() error
}

type pkiSetupRunner struct {
	runConfig pkiSetupRunConfig
}

type pkiSetupRunConfig struct {
	pkiSetupRunPath       string
	pkiSetupVaultJSONPath string
	resourceDirPath       string
}

var pkiSetupExector = newPkiSetupRunner()

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
	resourceDirPath := filepath.Join(baseWorkingDir, resourceDirName)

	scratchPath := filepath.Join(os.Getenv(envXdgRuntimeDir), pkiInitScratchDir)

	log.Println("pkiSetupRunPath: ", pkiSetupRunPath,
		"  pkiSetupVaultJSONPath: ", pkiSetupVaultJSONPath,
		"  scratchPath: ", scratchPath,
		"  resourceDirPath: ", resourceDirPath)

	pkiSetupExector.setRunConfig(pkiSetupRunConfig{
		pkiSetupRunPath:       pkiSetupRunPath,
		pkiSetupVaultJSONPath: pkiSetupVaultJSONPath,
		resourceDirPath:       resourceDirPath,
	})

	if err := pkiSetupExector.checkRunConfig(); err != nil {
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

	if err := pkiSetupExector.call(); err != nil {
		return exitWithError, err
	}

	return rearrangePkiByServices(pkiSetupVaultJSONPath)
}

func newPkiSetupRunner() pkiSetupRunnable {
	return &pkiSetupRunner{}
}

func (exect *pkiSetupRunner) setRunConfig(runConfig pkiSetupRunConfig) {
	exect.runConfig = runConfig
}

func (exect *pkiSetupRunner) checkRunConfig() error {
	if _, err := exec.LookPath(exect.runConfig.pkiSetupRunPath); err != nil {
		return err
	}

	if _, err := os.Stat(exect.runConfig.pkiSetupVaultJSONPath); os.IsNotExist(err) {
		log.Println("Vault JSON file for pkisetup not exists in ", exect.runConfig.pkiSetupVaultJSONPath)
		return err
	}

	return nil
}

func (exect *pkiSetupRunner) call() error {
	cmd := exec.Command(exect.runConfig.pkiSetupRunPath,
		"--config="+exect.runConfig.pkiSetupVaultJSONPath,
		"--confdir="+exect.runConfig.resourceDirPath)
	result, execErr := cmd.CombinedOutput()
	log.Printf("result of pkisetup: %s\n", string(result))
	return execErr
}

func rearrangePkiByServices(pkiSetupVaultJSONPath string) (exitCode, error) {
	config, readErr := config.NewX509Config(pkiSetupVaultJSONPath)
	if readErr != nil {
		return exitWithError, readErr
	}

	pkiOutputDir, err := config.PkiCADir()
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

func copyGeneratedForService(servicePath, pkiOutputDir string, config config.X509Config) error {
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
