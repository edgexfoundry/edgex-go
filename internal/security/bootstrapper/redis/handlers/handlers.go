/*******************************************************************************
* Copyright 2021 Intel Corporation
* Copyright 2020 Redis Labs
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
*
*******************************************************************************/

package handlers

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/helper"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/redis/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

// Handler is the redis bootstrapping handler
type Handler struct {
	credentials bootstrapConfig.Credentials
}

// NewHandler instantiates a new Handler
func NewHandler() *Handler {
	return &Handler{}
}

// GetCredentials retrieves the redis database credentials from secretstore
func (handler *Handler) GetCredentials(ctx context.Context, _ *sync.WaitGroup, startupTimer startup.Timer,
	dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)
	secretProvider := bootstrapContainer.SecretProviderFrom(dic.Get)

	var credentials = bootstrapConfig.Credentials{
		Username: "unset",
		Password: "unset",
	}

	for startupTimer.HasNotElapsed() {
		// retrieve database credentials from secretstore
		secrets, err := secretProvider.GetSecrets(config.Databases["Primary"].Type)
		if err == nil {
			credentials.Username = secrets[secret.UsernameKey]
			credentials.Password = secrets[secret.PasswordKey]
			break
		}

		lc.Warnf("Could not retrieve database credentials (startup timer has not expired): %s", err.Error())
		startupTimer.SleepForInterval()
	}

	if credentials.Password == "unset" {
		lc.Error("Failed to retrieve database credentials before startup timer expired")
		return false
	}

	handler.credentials = credentials
	return true
}

// SetupConfFile dynamically creates redis config file with the retrieved credentials
func (handler *Handler) SetupConfFile(ctx context.Context, _ *sync.WaitGroup, _ startup.Timer,
	dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)

	dbConfigDir := strings.TrimSpace(config.DatabaseConfig.Path)
	dbConfigFile := strings.TrimSpace(config.DatabaseConfig.Name)

	// required
	if dbConfigDir == "" {
		lc.Error("required configuration for DatabaseConfig.Path is empty")
		return false
	}

	if dbConfigFile == "" {
		lc.Error("required configuration for DatabaseConfig.Name is empty")
		return false
	}

	if err := helper.CreateDirectoryIfNotExists(dbConfigDir); err != nil {
		lc.Errorf("failed to create database config directory %s: %v", dbConfigDir, err)
		return false
	}

	dbConfigFilePath := filepath.Join(dbConfigDir, dbConfigFile)
	lc.Infof("Setting up the database config file %s", dbConfigFilePath)

	// open config file with read-write and overwritten attribute (TRUNC)
	confFile, err := os.OpenFile(dbConfigFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		lc.Errorf("failed to open db config file %s: %v", dbConfigFilePath, err)
		return false
	}
	defer func() {
		_ = confFile.Close()
	}()

	// writing the config file
	fwriter := bufio.NewWriter(confFile)
	if err := helper.GenerateConfig(fwriter, &handler.credentials.Password); err != nil {
		lc.Errorf("cannot write the db config file %s: %v", dbConfigFilePath, err)
		return false
	}
	if err := fwriter.Flush(); err != nil {
		lc.Errorf("failed to flush the file writer buffer %v", err)
		return false
	}

	lc.Info("database credentials have been set in the config file")

	return true
}
