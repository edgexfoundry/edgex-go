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
	"os"
	"path/filepath"

	x509 "github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/contract"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/helper"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

const (
	CommandName         = "generate"
	PkiSetupVaultJSON   = "pkisetup-vault.json"
	PkiSetupKongJSON    = "pkisetup-kong.json"
	PkiInitScratchDir   = "scratch"
	tlsCertFileName     = "server.crt"
	caCertFileName      = "ca.pem"
	PkiInitGeneratedDir = "generated"
	TlsSecretFileName   = "server.key"
	CaServiceName       = "ca"
	vaultServiceName    = "edgex-vault"
	kongServiceName     = "edgex-kong"
)

type Command struct {
	loggingClient logger.LoggingClient
	configuration *config.ConfigurationStruct
}

func NewCommand(
	loggingClient logger.LoggingClient,
	configuration *config.ConfigurationStruct) (*Command, *flag.FlagSet) {

	return &Command{
		loggingClient: loggingClient,
		configuration: configuration,
	}, flag.NewFlagSet(CommandName, flag.ExitOnError)
}

func (c *Command) Execute() (statusCode int, err error) {
	if statusCode, err := c.GeneratePkis(); err != nil {
		return statusCode, err
	}

	workDir, err := helper.GetWorkDir(c.configuration)
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}

	generatedDirPath := filepath.Join(workDir, PkiInitGeneratedDir)
	defer os.RemoveAll(generatedDirPath)

	// Shred the CA private key before deploy
	caPrivateKeyFile := filepath.Join(generatedDirPath, CaServiceName, TlsSecretFileName)
	if err := helper.SecureEraseFile(caPrivateKeyFile); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	deployDir, err := helper.GetDeployDir(c.configuration)
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}

	if err := helper.Deploy(generatedDirPath, deployDir, c.loggingClient); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	return contract.StatusCodeExitNormal, nil
}

func (c *Command) GeneratePkis() (int, error) {
	certConfigDir, err := helper.GetCertConfigDir(c.configuration)
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}

	certConfigDir, err = filepath.Abs(certConfigDir)
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}
	pkiSetupVaultJSONPath := filepath.Join(certConfigDir, PkiSetupVaultJSON)
	pkiSetupKongJSONPath := filepath.Join(certConfigDir, PkiSetupKongJSON)

	workingDir, err := helper.GetWorkDir(c.configuration)
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}
	scratchPath := filepath.Join(workingDir, PkiInitScratchDir)

	c.loggingClient.Debug(fmt.Sprint("pkiSetupVaultJSONPath: ", pkiSetupVaultJSONPath,
		"  pkiSetupKongJSONPath: ", pkiSetupKongJSONPath,
		"  scratchPath: ", scratchPath,
		"  certConfigDir: ", certConfigDir))

	if !helper.CheckIfFileExists(pkiSetupVaultJSONPath) {
		return contract.StatusCodeExitWithError, fmt.Errorf("Vault JSON file for security-secrets-setup does not exist in %s", pkiSetupVaultJSONPath)
	}

	if !helper.CheckIfFileExists(pkiSetupKongJSONPath) {
		return contract.StatusCodeExitWithError, fmt.Errorf("Kong JSON file for security-secrets-setup does not exist in %s", pkiSetupKongJSONPath)
	}

	// create scratch dir if not exists yet:
	if err := helper.CreateDirectoryIfNotExists(scratchPath); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	currDir, err := os.Getwd()
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}

	// after done, need to change it back to the original working dir to avoid os.Getwd() error
	// and delete the scratch dir
	defer c.cleanup(currDir, scratchPath)

	// generate TLS certs on the env. of $XDG_RUNTIME_DIR/edgex/pki-init/scratch
	if err := os.Chdir(scratchPath); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	if err := helper.GenTLSAssets(pkiSetupVaultJSONPath, c.loggingClient); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	if err := helper.GenTLSAssets(pkiSetupKongJSONPath, c.loggingClient); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	return c.rearrangePkiByServices(workingDir, pkiSetupVaultJSONPath, pkiSetupKongJSONPath)
}

func (c *Command) rearrangePkiByServices(workingDir, pkiSetupVaultJSONPath, pkiSetupKongJSONPath string) (int, error) {
	vaultConfig, readErr := x509.NewX509Config(pkiSetupVaultJSONPath)
	if readErr != nil {
		return contract.StatusCodeExitWithError, readErr
	}

	kongConfig, readErr := x509.NewX509Config(pkiSetupKongJSONPath)
	if readErr != nil {
		return contract.StatusCodeExitWithError, readErr
	}

	generatedDirPath := filepath.Join(workingDir, PkiInitGeneratedDir)

	c.loggingClient.Debug(fmt.Sprint("pki-init generate output base dir: ", generatedDirPath))

	// create generated dir if not exists yet:
	if err := helper.CreateDirectoryIfNotExists(generatedDirPath); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	// CA:
	caDirPath := filepath.Join(generatedDirPath, CaServiceName)
	if err := c.copyGeneratedForService(caDirPath, vaultConfig); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	// Vault:
	vaultServicePath := filepath.Join(generatedDirPath, vaultServiceName)
	if err := c.copyGeneratedForService(vaultServicePath, vaultConfig); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	// Kong:
	kongServicePath := filepath.Join(generatedDirPath, kongServiceName)
	if err := c.copyGeneratedForService(kongServicePath, kongConfig); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	return contract.StatusCodeExitNormal, nil
}

func (c *Command) copyGeneratedForService(servicePath string, config x509.X509Config) error {
	if err := helper.CreateDirectoryIfNotExists(servicePath); err != nil {
		return err
	}

	pkiOutputDir, err := config.PkiCADir()
	if err != nil {
		return err
	}

	if _, err := helper.CopyFile(filepath.Join(pkiOutputDir, config.GetCAPemFileName()), filepath.Join(servicePath, caCertFileName)); err != nil {
		return err
	}

	privKeyFileName := filepath.Join(servicePath, TlsSecretFileName)
	if filepath.Base(servicePath) == CaServiceName {
		if _, err := helper.CopyFile(filepath.Join(pkiOutputDir, config.GetCAPrivateKeyFileName()), privKeyFileName); err != nil {
			return err
		}
	} else {
		if _, err := helper.CopyFile(filepath.Join(pkiOutputDir, config.GetTLSPrivateKeyFileName()), privKeyFileName); err != nil {
			return err
		}
		// if not CA, then also copy the TLS cert as well
		if _, err := helper.CopyFile(filepath.Join(pkiOutputDir, config.GetTLSPemFileName()), filepath.Join(servicePath, tlsCertFileName)); err != nil {
			return err
		}
	}

	// read-only to the owner
	return os.Chmod(privKeyFileName, 0400)
}

func (c *Command) cleanup(origWorkingDir, scratchPath string) {
	_ = os.Chdir(origWorkingDir)
	os.RemoveAll(scratchPath)
	c.loggingClient.Info("pki-init generation completes")
}
