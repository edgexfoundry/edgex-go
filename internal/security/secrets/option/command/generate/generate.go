/*******************************************************************************
 * Copyright 2019 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package generate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option"

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

func (g *Command) Execute() (statusCode int, err error) {
	if statusCode, err := g.GeneratePkis(); err != nil {
		return statusCode, err
	}

	workDir, err := option.GetWorkDir()
	if err != nil {
		return option.ExitWithError, err
	}

	generatedDirPath := filepath.Join(workDir, option.PkiInitGeneratedDir)
	defer os.RemoveAll(generatedDirPath)

	// Shred the CA private key before deploy
	caPrivateKeyFile := filepath.Join(generatedDirPath, option.CaServiceName, option.TlsSecretFileName)
	if err := option.SecureEraseFile(caPrivateKeyFile); err != nil {
		return option.ExitWithError, err
	}

	deployDir, err := option.GetDeployDir()
	if err != nil {
		return option.ExitWithError, err
	}

	if err := option.Deploy(generatedDirPath, deployDir); err != nil {
		return option.ExitWithError, err
	}

	return option.ExitNormal, nil
}

func (g *Command) GeneratePkis() (int, error) {
	certConfigDir, err := option.GetCertConfigDir()
	if err != nil {
		return option.ExitWithError, err
	}

	certConfigDir, err = filepath.Abs(certConfigDir)
	if err != nil {
		return option.ExitWithError, err
	}
	pkiSetupVaultJSONPath := filepath.Join(certConfigDir, option.PkiSetupVaultJSON)
	pkiSetupKongJSONPath := filepath.Join(certConfigDir, option.PkiSetupKongJSON)

	workingDir, err := option.GetWorkDir()
	if err != nil {
		return option.ExitWithError, err
	}
	scratchPath := filepath.Join(workingDir, option.PkiInitScratchDir)

	g.loggingClient.Debug(fmt.Sprint("pkiSetupVaultJSONPath: ", pkiSetupVaultJSONPath,
		"  pkiSetupKongJSONPath: ", pkiSetupKongJSONPath,
		"  scratchPath: ", scratchPath,
		"  certConfigDir: ", certConfigDir))

	if !option.CheckIfFileExists(pkiSetupVaultJSONPath) {
		return option.ExitWithError, fmt.Errorf("Vault JSON file for security-secrets-setup does not exist in %s", pkiSetupVaultJSONPath)
	}

	if !option.CheckIfFileExists(pkiSetupKongJSONPath) {
		return option.ExitWithError, fmt.Errorf("Kong JSON file for security-secrets-setup does not exist in %s", pkiSetupKongJSONPath)
	}

	// create scratch dir if not exists yet:
	if err := option.CreateDirectoryIfNotExists(scratchPath); err != nil {
		return option.ExitWithError, err
	}

	currDir, err := os.Getwd()
	if err != nil {
		return option.ExitWithError, err
	}

	// after done, need to change it back to the original working dir to avoid os.Getwd() error
	// and delete the scratch dir
	defer g.cleanup(currDir, scratchPath)

	// generate TLS certs on the env. of $XDG_RUNTIME_DIR/edgex/pki-init/scratch
	if err := os.Chdir(scratchPath); err != nil {
		return option.ExitWithError, err
	}

	if err := option.GenTLSAssets(pkiSetupVaultJSONPath); err != nil {
		return option.ExitWithError, err
	}

	if err := option.GenTLSAssets(pkiSetupKongJSONPath); err != nil {
		return option.ExitWithError, err
	}

	return g.rearrangePkiByServices(workingDir, pkiSetupVaultJSONPath, pkiSetupKongJSONPath)
}

func (g *Command) rearrangePkiByServices(workingDir, pkiSetupVaultJSONPath, pkiSetupKongJSONPath string) (int, error) {
	vaultConfig, readErr := config.NewX509Config(pkiSetupVaultJSONPath)
	if readErr != nil {
		return option.ExitWithError, readErr
	}

	kongConfig, readErr := config.NewX509Config(pkiSetupKongJSONPath)
	if readErr != nil {
		return option.ExitWithError, readErr
	}

	generatedDirPath := filepath.Join(workingDir, option.PkiInitGeneratedDir)

	g.loggingClient.Debug(fmt.Sprint("pki-init generate output base dir: ", generatedDirPath))

	// create generated dir if not exists yet:
	if err := option.CreateDirectoryIfNotExists(generatedDirPath); err != nil {
		return option.ExitWithError, err
	}

	// CA:
	caDirPath := filepath.Join(generatedDirPath, option.CaServiceName)
	if err := g.copyGeneratedForService(caDirPath, vaultConfig); err != nil {
		return option.ExitWithError, err
	}

	// Vault:
	vaultServicePath := filepath.Join(generatedDirPath, option.VaultServiceName)
	if err := g.copyGeneratedForService(vaultServicePath, vaultConfig); err != nil {
		return option.ExitWithError, err
	}

	// Kong:
	kongServicePath := filepath.Join(generatedDirPath, option.KongServiceName)
	if err := g.copyGeneratedForService(kongServicePath, kongConfig); err != nil {
		return option.ExitWithError, err
	}

	return option.ExitNormal, nil
}

func (g *Command) copyGeneratedForService(servicePath string, config config.X509Config) error {
	if err := option.CreateDirectoryIfNotExists(servicePath); err != nil {
		return err
	}

	pkiOutputDir, err := config.PkiCADir()
	if err != nil {
		return err
	}

	if _, err := option.CopyFile(filepath.Join(pkiOutputDir, config.GetCAPemFileName()), filepath.Join(servicePath, option.CaCertFileName)); err != nil {
		return err
	}

	privKeyFileName := filepath.Join(servicePath, option.TlsSecretFileName)
	if filepath.Base(servicePath) == option.CaServiceName {
		if _, err := option.CopyFile(filepath.Join(pkiOutputDir, config.GetCAPrivateKeyFileName()), privKeyFileName); err != nil {
			return err
		}
	} else {
		if _, err := option.CopyFile(filepath.Join(pkiOutputDir, config.GetTLSPrivateKeyFileName()), privKeyFileName); err != nil {
			return err
		}
		// if not CA, then also copy the TLS cert as well
		if _, err := option.CopyFile(filepath.Join(pkiOutputDir, config.GetTLSPemFileName()), filepath.Join(servicePath, option.TlsCertFileName)); err != nil {
			return err
		}
	}

	// read-only to the owner
	return os.Chmod(privKeyFileName, 0400)
}

func (g *Command) cleanup(origWorkingDir, scratchPath string) {
	_ = os.Chdir(origWorkingDir)
	os.RemoveAll(scratchPath)
	g.loggingClient.Info("pki-init generation completes")
}
