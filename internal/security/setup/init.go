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

package setup

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

// Global variables
var Configuration *ConfigurationStruct

func Init() {
	// Unfortunately I have to do this because of utilization of the LoggingClient in config.LoadFromFile below.
	// That function is expecting a LoggingClient instance. Right now, it appears that is only set via global var in that package
	// TODO: This doesn't make any sense. Review this usage as applicable to all other service init routines.
	//       See TODO in internal/pkg/config/loader.go where var LoggingClient is declared.
	lc := logger.NewClient("edgex-security-pkisetup", false, "", models.InfoLog)
	config.LoggingClient = lc

	var err error
	Configuration, err = initializeConfiguration(false, "") //These values are defaults. Preserved variables for possible later extension
	if err != nil {
		lc.Error(err.Error())
	}
}

func initializeConfiguration(useRegistry bool, useProfile string) (*ConfigurationStruct, error) {
	// We currently have to load configuration from filesystem first in order to obtain Registry Host/Port
	configuration := &ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, configuration)
	if err != nil {
		return nil, err
	}

	return configuration, nil
}
