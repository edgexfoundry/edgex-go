/********************************************************************************
 *  Copyright 2019 Dell Inc.
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

package handlers

import (
	"context"
	"os"
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-registry/registry"
	"github.com/edgexfoundry/go-mod-secrets/pkg"
	"github.com/edgexfoundry/go-mod-secrets/pkg/providers/vault"

	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
)

// Environment variable used to disable the secret store functionality which is ON/TRUE by default.
const SecretStore = "EDGEX_SECURITY_SECRET_STORE"

// SecretClient is a global variable used to house the SecretClient to be used system-wide for obtaining secrets.
var SecretClient pkg.SecretClient

// SecretClientBootstrapHandler creates a SecretClient to be used for obtaining secrets from a secret store manager.
// NOTE: This BootstrapHandler is responsible for creating a utility that will most likely be used by other
// BootstrapHandlers to obtain sensitive data, such as database credentials. This BootstrapHandler should be processed
// before other BootstrapHandlers, possibly even first since it has not other dependencies.
func SecretClientBootstrapHandler(
	wg *sync.WaitGroup,
	context context.Context,
	startupTimer startup.Timer,
	config interfaces.Configuration,
	loggingClient logger.LoggingClient,
	registryClient registry.Client) bool {

	if env := os.Getenv(SecretStore); env != "false" {
		return true
	}

	var err error
	SecretClient, err = vault.NewSecretClient(config.GetBootstrap().SecretStore)

	return err == nil
}
