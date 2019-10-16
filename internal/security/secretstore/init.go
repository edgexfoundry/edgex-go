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
 *
 *******************************************************************************/

package secretstore

import (
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

// Global variables
var Configuration *ConfigurationStruct
var LoggingClient logger.LoggingClient

func Retry(useRegistry bool, configDir, profileDir string, timeout int, wait *sync.WaitGroup, ch chan error) {
	until := time.Now().Add(time.Millisecond * time.Duration(timeout))
	for time.Now().Before(until) {
		var err error
		// When looping, only handle configuration if it hasn't already been set.
		if Configuration == nil {
			// Next two lines are workaround for issue #1814 (nil panic while logging)
			// where config.LoadFromFile depends on a global LoggingClient that isn't set anywhere
			// Remove this workaround once this tool is migrated to common bootstrap.
			lc := logger.NewClient(internal.SecuritySecretStoreSetupServiceKey, false, "", models.InfoLog)
			config.LoggingClient = lc
			Configuration, err = initializeConfiguration(useRegistry, configDir, profileDir)
			if err != nil {
				ch <- err
				if !useRegistry {
					// Error occurred when attempting to read from local filesystem. Fail fast.
					close(ch)
					wait.Done()
					return
				}
			} else {
				// Setup Logging
				logTarget := setLoggingTarget()
				LoggingClient = logger.NewClient(internal.SecuritySecretStoreSetupServiceKey, Configuration.Logging.EnableRemote, logTarget, Configuration.Writable.LogLevel)
			}
		}
		// This seems a bit artificial here due to lack of additional service requirements
		// but conforms to the pattern found in other edgex-go services.
		if Configuration != nil {
			break
		}
		time.Sleep(time.Second * time.Duration(1))
	}
	close(ch)
	wait.Done()

	return
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

func setLoggingTarget() string {
	return Configuration.Logging.File
}
