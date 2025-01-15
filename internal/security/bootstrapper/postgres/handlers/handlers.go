//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/helper"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/postgres/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
)

const (
	postgresSecretName = "postgres"
	passwordFileDir    = "/run/secrets"
	passwordFileName   = "postgres_password"
)

// SetupDBScriptFiles dynamically creates Postgres init-db script file with the retrieved credentials for multiple EdgeX services
func SetupDBScriptFiles(_ context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)

	dbInitScriptDir := strings.TrimSpace(config.DatabaseConfig.Path)
	dbScriptFile := strings.TrimSpace(config.DatabaseConfig.Name)

	// required
	if dbInitScriptDir == "" {
		lc.Error("required configuration for DatabaseConfig.Path is empty")
		return false
	}

	if dbScriptFile == "" {
		lc.Error("required configuration for DatabaseConfig.Name is empty")
		return false
	}

	if err := helper.CreateDirectoryIfNotExists(dbInitScriptDir); err != nil {
		lc.Errorf("failed to create database initialized script directory %s: %v", dbInitScriptDir, err)
		return false
	}

	// create Postgres init-db script file
	confFile, err := helper.CreateConfigFile(dbInitScriptDir, dbScriptFile, lc)
	if err != nil {
		lc.Error(err.Error())
		return false
	}
	defer func() {
		_ = confFile.Close()
	}()

	err = getServiceCredentials(dic, confFile)
	if err != nil {
		lc.Error(err.Error())
		return false
	}
	return true
}

// getServiceCredentials retrieves the Postgres database credentials of multiple services from secretstore
func getServiceCredentials(dic *di.Container, scriptFile *os.File) error {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	secretProvider := bootstrapContainer.SecretProviderFrom(dic.Get)

	secretList, err := secretProvider.ListSecretNames()
	if err != nil {
		return err
	} else {
		var credMap []map[string]any
		for _, serviceKey := range secretList {
			exists, err := secretProvider.HasSecret(path.Join(serviceKey, postgresSecretName))
			if err != nil {
				return err
			}
			if exists {
				serviceSecrets, err := secretProvider.GetSecret(path.Join(serviceKey, postgresSecretName))
				if err != nil {
					return err
				}
				username, userFound := serviceSecrets[secret.UsernameKey]
				password, pwFound := serviceSecrets[secret.PasswordKey]
				if userFound && pwFound {
					dbCred := map[string]any{
						helper.UsernameTempVarName: username,
						helper.PasswordTempVarName: &password,
					}
					credMap = append(credMap, dbCred)
				}
			}
		}

		// writing the Postgres init-db script with the Postgres credentials got from secret store
		if err := helper.GeneratePostgresScript(scriptFile, credMap); err != nil {
			lc.Errorf("cannot write to init-db script file %s: %v", scriptFile.Name(), err)
			return err
		}

		lc.Info("Postgres init-db script has been set")
	}
	return nil
}

// SetupPasswordFile creates the Postgres superuser password file with the credential retrieved from secret provider
func SetupPasswordFile(_ context.Context, _ *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)

	if err := helper.CreateDirectoryIfNotExists(passwordFileDir); err != nil {
		lc.Errorf("failed to create database superuser password file directory %s: %v", passwordFileDir, err)
		return false
	}

	// Create the Postgres superuser password file
	confFile, err := helper.CreateConfigFile(passwordFileDir, passwordFileName, lc)
	if err != nil {
		lc.Error(err.Error())
		return false
	}
	defer func() {
		_ = confFile.Close()
	}()

	// GetCredentials retrieves the Postgres database credentials from secretstore
	secretProvider := bootstrapContainer.SecretProviderFrom(dic.Get)

	var superuserPass string

	for startupTimer.HasNotElapsed() {
		// retrieve database credentials from secretstore
		secrets, err := secretProvider.GetSecret(config.Database.Type)
		if err == nil {
			superuserPass = secrets[secret.PasswordKey]
			break
		}

		lc.Warnf("Could not retrieve database credentials (startup timer has not expired): %s", err.Error())
		startupTimer.SleepForInterval()
	}

	if superuserPass == "" {
		lc.Error("Failed to retrieve database credentials before startup timer expired")
		return false
	}

	// Writing the Postgres password file with the Postgres credentials got from secret store
	if genErr := helper.GeneratePasswordFile(confFile, superuserPass); genErr != nil {
		lc.Errorf("cannot write password to file %s: %v", passwordFileName, genErr)
		return false
	}

	lc.Info("Postgres password file has been set")
	return true
}
