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
	"context"
	"fmt"
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
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
)

const (
	// redis aclfile name
	redisACLFileName = "edgex_redis_acl.conf"

	// file read/write permission only for owner
	readWriteOnlyForOwner os.FileMode = 0600
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
		secrets, err := secretProvider.GetSecret(config.Databases[common.Primary].Type)
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

// SetupConfFiles dynamically creates redis config file with the retrieved credentials and setup ACLs
func (handler *Handler) SetupConfFiles(ctx context.Context, _ *sync.WaitGroup, _ startup.Timer,
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

	// create redis config file
	confFile, err := createConfigFile(dbConfigDir, dbConfigFile, lc)
	if err != nil {
		lc.Error(err.Error())
		return false
	}
	defer func() {
		_ = confFile.Close()
	}()

	edgeXRedisACLFilePath := filepath.Join(dbConfigDir, redisACLFileName)
	// writing the config file
	if err := helper.GenerateRedisConfig(confFile, edgeXRedisACLFilePath); err != nil {
		lc.Errorf("cannot write the db config file %s: %v", confFile.Name(), err)
		return false
	}

	// create ACL config file
	aclFile, err := createConfigFile(dbConfigDir, redisACLFileName, lc)
	if err != nil {
		lc.Error(err.Error())
		return false
	}
	defer func() {
		_ = aclFile.Close()
	}()

	// write the ACL file
	if err := helper.GenerateACLConfig(aclFile, &handler.credentials.Password); err != nil {
		lc.Errorf("cannot write the ACL config file %s: %v", edgeXRedisACLFilePath, err)
		return false
	}

	lc.Info("database config and ACL have been set in the config files")

	return true
}

func createConfigFile(configDir, configFileName string, lc logger.LoggingClient) (*os.File, error) {
	configFilePath := filepath.Join(configDir, configFileName)
	lc.Infof("Creating the database config file: %s", configFilePath)

	// open config file with read-write and overwritten attribute (TRUNC)
	configFile, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, readWriteOnlyForOwner)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file %s: %v", configFilePath, err)
	}
	return configFile, nil
}
