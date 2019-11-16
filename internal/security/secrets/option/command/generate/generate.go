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

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/contract"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/helper"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

const (
	CommandGenerate     = "generate"
	pkiSetupVaultJSON   = "pkisetup-vault.json"
	pkiSetupKongJSON    = "pkisetup-kong.json"
	pkiInitScratchDir   = "scratch"
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
	helper        *helper.Helper
}

func NewCommand(flags *FlagSet, loggingClient logger.LoggingClient, helper *helper.Helper) (*Command, *flag.FlagSet) {
	return &Command{
			loggingClient: loggingClient,
			helper:        helper,
		},
		flags.flagSet
}

func (g *Command) Execute() (statusCode int, err error) {
	if statusCode, err := g.GeneratePkis(); err != nil {
		return statusCode, err
	}

	workDir, err := g.helper.GetWorkDir()
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}

	generatedDirPath := filepath.Join(workDir, PkiInitGeneratedDir)
	defer os.RemoveAll(generatedDirPath)

	// Shred the CA private key before deploy
	caPrivateKeyFile := filepath.Join(generatedDirPath, CaServiceName, TlsSecretFileName)
	if err := g.helper.SecureEraseFile(caPrivateKeyFile); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	deployDir, err := g.helper.GetDeployDir()
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}

	if err := g.helper.Deploy(generatedDirPath, deployDir); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	return contract.StatusCodeExitNormal, nil
}

func (g *Command) GeneratePkis() (int, error) {
	certConfigDir, err := g.helper.GetCertConfigDir()
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}

	certConfigDir, err = filepath.Abs(certConfigDir)
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}
	pkiSetupVaultJSONPath := filepath.Join(certConfigDir, pkiSetupVaultJSON)
	pkiSetupKongJSONPath := filepath.Join(certConfigDir, pkiSetupKongJSON)

	workingDir, err := g.helper.GetWorkDir()
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}
	scratchPath := filepath.Join(workingDir, pkiInitScratchDir)

	g.loggingClient.Debug(fmt.Sprint("pkiSetupVaultJSONPath: ", pkiSetupVaultJSONPath,
		"  pkiSetupKongJSONPath: ", pkiSetupKongJSONPath,
		"  scratchPath: ", scratchPath,
		"  certConfigDir: ", certConfigDir))

	if !g.helper.CheckIfFileExists(pkiSetupVaultJSONPath) {
		return contract.StatusCodeExitWithError, fmt.Errorf("Vault JSON file for security-secrets-setup does not exist in %s", pkiSetupVaultJSONPath)
	}

	if !g.helper.CheckIfFileExists(pkiSetupKongJSONPath) {
		return contract.StatusCodeExitWithError, fmt.Errorf("Kong JSON file for security-secrets-setup does not exist in %s", pkiSetupKongJSONPath)
	}

	// create scratch dir if not exists yet:
	if err := g.helper.CreateDirectoryIfNotExists(scratchPath); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	currDir, err := os.Getwd()
	if err != nil {
		return contract.StatusCodeExitWithError, err
	}

	// after done, need to change it back to the original working dir to avoid os.Getwd() error
	// and delete the scratch dir
	defer g.cleanup(currDir, scratchPath)

	// generate TLS certs on the env. of $XDG_RUNTIME_DIR/edgex/pki-init/scratch
	if err := os.Chdir(scratchPath); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	if err := g.helper.GenTLSAssets(pkiSetupVaultJSONPath); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	if err := g.helper.GenTLSAssets(pkiSetupKongJSONPath); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	return g.rearrangePkiByServices(workingDir, pkiSetupVaultJSONPath, pkiSetupKongJSONPath)
}

func (g *Command) rearrangePkiByServices(workingDir, pkiSetupVaultJSONPath, pkiSetupKongJSONPath string) (int, error) {
	vaultConfig, readErr := config.NewX509Config(pkiSetupVaultJSONPath)
	if readErr != nil {
		return contract.StatusCodeExitWithError, readErr
	}

	kongConfig, readErr := config.NewX509Config(pkiSetupKongJSONPath)
	if readErr != nil {
		return contract.StatusCodeExitWithError, readErr
	}

	generatedDirPath := filepath.Join(workingDir, PkiInitGeneratedDir)

	g.loggingClient.Debug(fmt.Sprint("pki-init generate output base dir: ", generatedDirPath))

	// create generated dir if not exists yet:
	if err := g.helper.CreateDirectoryIfNotExists(generatedDirPath); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	// CA:
	caDirPath := filepath.Join(generatedDirPath, CaServiceName)
	if err := g.copyGeneratedForService(caDirPath, vaultConfig); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	// Vault:
	vaultServicePath := filepath.Join(generatedDirPath, vaultServiceName)
	if err := g.copyGeneratedForService(vaultServicePath, vaultConfig); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	// Kong:
	kongServicePath := filepath.Join(generatedDirPath, kongServiceName)
	if err := g.copyGeneratedForService(kongServicePath, kongConfig); err != nil {
		return contract.StatusCodeExitWithError, err
	}

	return contract.StatusCodeExitNormal, nil
}

func (g *Command) copyGeneratedForService(servicePath string, config config.X509Config) error {
	if err := g.helper.CreateDirectoryIfNotExists(servicePath); err != nil {
		return err
	}

	pkiOutputDir, err := config.PkiCADir()
	if err != nil {
		return err
	}

	if _, err := g.helper.CopyFile(filepath.Join(pkiOutputDir, config.GetCAPemFileName()), filepath.Join(servicePath, caCertFileName)); err != nil {
		return err
	}

	privKeyFileName := filepath.Join(servicePath, TlsSecretFileName)
	if filepath.Base(servicePath) == CaServiceName {
		if _, err := g.helper.CopyFile(filepath.Join(pkiOutputDir, config.GetCAPrivateKeyFileName()), privKeyFileName); err != nil {
			return err
		}
	} else {
		if _, err := g.helper.CopyFile(filepath.Join(pkiOutputDir, config.GetTLSPrivateKeyFileName()), privKeyFileName); err != nil {
			return err
		}
		// if not CA, then also copy the TLS cert as well
		if _, err := g.helper.CopyFile(filepath.Join(pkiOutputDir, config.GetTLSPemFileName()), filepath.Join(servicePath, tlsCertFileName)); err != nil {
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
