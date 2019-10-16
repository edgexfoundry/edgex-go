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

package secrets

import (
	"fmt"
	"path/filepath"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

// Global variables
var Configuration *ConfigurationStruct
var LoggingClient logger.LoggingClient

func Init(configDir string) error {
	// Unfortunately I have to do this because of utilization of the LoggingClient in config.LoadFromFile below.
	// That function is expecting a LoggingClient instance. Right now, it appears that is only set via global var in that package
	// TODO: This doesn't make any sense. Review this usage as applicable to all other service init routines.
	//       See TODO in internal/pkg/config/loader.go where var LoggingClient is declared.
	lc := logger.NewClient(internal.SecuritySecretsSetupServiceKey, false, "", models.InfoLog)
	config.LoggingClient = lc

	var err error
	Configuration, err = initializeConfiguration(false, configDir, "") //These values are defaults. Preserved variables for possible later extension
	if err != nil {
		lc.Error(err.Error())
		return err
	}

	loggerAbsPath, err := filepath.Abs(Configuration.Logging.File)
	if err != nil {
		lc.Error(fmt.Sprintf("Error on finding the absolute path for logging client from configuration Logging.File file %s: %v\n", Configuration.Logging.File, err))
		return err
	}

	LoggingClient = logger.NewClient(internal.SecuritySecretsSetupServiceKey, Configuration.Logging.EnableRemote,
		loggerAbsPath, Configuration.Writable.LogLevel)

	return nil
}

func initializeConfiguration(useRegistry bool, configDir, profileDir string) (*ConfigurationStruct, error) {
	// We currently have to load configuration from filesystem first in order to obtain Registry Host/Port
	configuration := &ConfigurationStruct{}
	err := config.LoadFromFile(configDir, profileDir, configuration)
	if err != nil {
		return nil, err
	}

	return configuration, nil
}
