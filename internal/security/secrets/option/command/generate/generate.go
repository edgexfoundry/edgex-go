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
	"flag"
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/helper"
	"os"
	"path/filepath"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/constant"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

type Command struct {
	loggingClient logger.LoggingClient
}

func NewCommand(flags *FlagSet, loggingClient logger.LoggingClient) (*Command, *flag.FlagSet) {
	return &Command{
			loggingClient: loggingClient,
		},
		flags.flagSet
}

func (g *Command) Execute() (statusCode int, err error) {
	if statusCode, err := g.GeneratePkis(); err != nil {
		return statusCode, err
	}

	workDir, err := helper.GetWorkDir()
	if err != nil {
		return constant.ExitWithError, err
	}

	generatedDirPath := filepath.Join(workDir, constant.PkiInitGeneratedDir)
	defer os.RemoveAll(generatedDirPath)

	// Shred the CA private key before deploy
	caPrivateKeyFile := filepath.Join(generatedDirPath, constant.CaServiceName, constant.TlsSecretFileName)
	if err := helper.SecureEraseFile(caPrivateKeyFile); err != nil {
		return constant.ExitWithError, err
	}

	deployDir, err := helper.GetDeployDir()
	if err != nil {
		return constant.ExitWithError, err
	}

	if err := helper.Deploy(generatedDirPath, deployDir); err != nil {
		return constant.ExitWithError, err
	}

	return constant.ExitNormal, nil
}

func (g *Command) GeneratePkis() (int, error) {
	certConfigDir, err := helper.GetCertConfigDir()
	if err != nil {
		return constant.ExitWithError, err
	}

	certConfigDir, err = filepath.Abs(certConfigDir)
	if err != nil {
		return constant.ExitWithError, err
	}
	pkiSetupVaultJSONPath := filepath.Join(certConfigDir, constant.PkiSetupVaultJSON)
	pkiSetupKongJSONPath := filepath.Join(certConfigDir, constant.PkiSetupKongJSON)

	workingDir, err := helper.GetWorkDir()
	if err != nil {
		return constant.ExitWithError, err
	}
	scratchPath := filepath.Join(workingDir, constant.PkiInitScratchDir)

	g.loggingClient.Debug(fmt.Sprint("pkiSetupVaultJSONPath: ", pkiSetupVaultJSONPath,
		"  pkiSetupKongJSONPath: ", pkiSetupKongJSONPath,
		"  scratchPath: ", scratchPath,
		"  certConfigDir: ", certConfigDir))

	if !helper.CheckIfFileExists(pkiSetupVaultJSONPath) {
		return constant.ExitWithError, fmt.Errorf("Vault JSON file for security-secrets-setup does not exist in %s", pkiSetupVaultJSONPath)
	}

	if !helper.CheckIfFileExists(pkiSetupKongJSONPath) {
		return constant.ExitWithError, fmt.Errorf("Kong JSON file for security-secrets-setup does not exist in %s", pkiSetupKongJSONPath)
	}

	// create scratch dir if not exists yet:
	if err := helper.CreateDirectoryIfNotExists(scratchPath); err != nil {
		return constant.ExitWithError, err
	}

	currDir, err := os.Getwd()
	if err != nil {
		return constant.ExitWithError, err
	}

	// after done, need to change it back to the original working dir to avoid os.Getwd() error
	// and delete the scratch dir
	defer g.cleanup(currDir, scratchPath)

	// generate TLS certs on the env. of $XDG_RUNTIME_DIR/edgex/pki-init/scratch
	if err := os.Chdir(scratchPath); err != nil {
		return constant.ExitWithError, err
	}

	if err := helper.GenTLSAssets(pkiSetupVaultJSONPath); err != nil {
		return constant.ExitWithError, err
	}

	if err := helper.GenTLSAssets(pkiSetupKongJSONPath); err != nil {
		return constant.ExitWithError, err
	}

	return g.rearrangePkiByServices(workingDir, pkiSetupVaultJSONPath, pkiSetupKongJSONPath)
}

func (g *Command) rearrangePkiByServices(workingDir, pkiSetupVaultJSONPath, pkiSetupKongJSONPath string) (int, error) {
	vaultConfig, readErr := config.NewX509Config(pkiSetupVaultJSONPath)
	if readErr != nil {
		return constant.ExitWithError, readErr
	}

	kongConfig, readErr := config.NewX509Config(pkiSetupKongJSONPath)
	if readErr != nil {
		return constant.ExitWithError, readErr
	}

	generatedDirPath := filepath.Join(workingDir, constant.PkiInitGeneratedDir)

	g.loggingClient.Debug(fmt.Sprint("pki-init generate output base dir: ", generatedDirPath))

	// create generated dir if not exists yet:
	if err := helper.CreateDirectoryIfNotExists(generatedDirPath); err != nil {
		return constant.ExitWithError, err
	}

	// CA:
	caDirPath := filepath.Join(generatedDirPath, constant.CaServiceName)
	if err := g.copyGeneratedForService(caDirPath, vaultConfig); err != nil {
		return constant.ExitWithError, err
	}

	// Vault:
	vaultServicePath := filepath.Join(generatedDirPath, constant.VaultServiceName)
	if err := g.copyGeneratedForService(vaultServicePath, vaultConfig); err != nil {
		return constant.ExitWithError, err
	}

	// Kong:
	kongServicePath := filepath.Join(generatedDirPath, constant.KongServiceName)
	if err := g.copyGeneratedForService(kongServicePath, kongConfig); err != nil {
		return constant.ExitWithError, err
	}

	return constant.ExitNormal, nil
}

func (g *Command) copyGeneratedForService(servicePath string, config config.X509Config) error {
	if err := helper.CreateDirectoryIfNotExists(servicePath); err != nil {
		return err
	}

	pkiOutputDir, err := config.PkiCADir()
	if err != nil {
		return err
	}

	if _, err := helper.CopyFile(filepath.Join(pkiOutputDir, config.GetCAPemFileName()), filepath.Join(servicePath, constant.CaCertFileName)); err != nil {
		return err
	}

	privKeyFileName := filepath.Join(servicePath, constant.TlsSecretFileName)
	if filepath.Base(servicePath) == constant.CaServiceName {
		if _, err := helper.CopyFile(filepath.Join(pkiOutputDir, config.GetCAPrivateKeyFileName()), privKeyFileName); err != nil {
			return err
		}
	} else {
		if _, err := helper.CopyFile(filepath.Join(pkiOutputDir, config.GetTLSPrivateKeyFileName()), privKeyFileName); err != nil {
			return err
		}
		// if not CA, then also copy the TLS cert as well
		if _, err := helper.CopyFile(filepath.Join(pkiOutputDir, config.GetTLSPemFileName()), filepath.Join(servicePath, constant.TlsCertFileName)); err != nil {
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
