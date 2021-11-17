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

package handlers

import (
	"context"
	"fmt"
	"sync"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

// StartMessage contains references to dependencies required by the start message handler.
type StartMessage struct {
	serviceKey string
	version    string
}

// NewStartMessage is a factory method that returns an initialized StartMessage receiver struct.
func NewStartMessage(serviceKey, version string) *StartMessage {
	return &StartMessage{
		serviceKey: serviceKey,
		version:    version,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract.  It creates no go routines.  It logs a "standard" set of
// messages when the service first starts up successfully.
func (h StartMessage) BootstrapHandler(
	_ context.Context,
	_ *sync.WaitGroup,
	startupTimer startup.Timer,
	dic *di.Container) bool {

	lc := container.LoggingClientFrom(dic.Get)
	lc.Info("Service dependencies resolved...")
	lc.Info(fmt.Sprintf("Starting %s %s ", h.serviceKey, h.version))

	bootstrapConfig := container.ConfigurationFrom(dic.Get).GetBootstrap()
	if len(bootstrapConfig.Service.StartupMsg) > 0 {
		lc.Info(bootstrapConfig.Service.StartupMsg)
	}

	lc.Info("Service started in: " + startupTimer.SinceAsString())

	return true
}
