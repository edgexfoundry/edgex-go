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

package logging

import (
	"path/filepath"

	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

// FactoryToStdout returns a logger.LoggingClient that outputs to stdout.
func FactoryToStdout(serviceKey string) logger.LoggingClient {
	return logger.NewClientStdOut(serviceKey, false, models.DebugLog)
}

// FactoryFromConfiguration returns a logger.LoggingClient based on configuration settings.
func FactoryFromConfiguration(serviceKey string, config interfaces.Configuration) logger.LoggingClient {
	var target string
	bootstrapConfig := config.GetBootstrap()
	if bootstrapConfig.Logging.EnableRemote {
		target = bootstrapConfig.Clients["Logging"].Url() + clients.ApiLoggingRoute
	} else {
		target, _ = filepath.Abs(bootstrapConfig.Logging.File)
	}
	return logger.NewClient(serviceKey, bootstrapConfig.Logging.EnableRemote, target, config.GetLogLevel())
}
